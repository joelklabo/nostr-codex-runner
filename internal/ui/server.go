package ui

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"nostr-codex-runner/internal/config"
)

//go:embed web/*
var embeddedFS embed.FS

// Server hosts the local web UI and bd-backed APIs.
type Server struct {
	cfg    *config.Config
	logger Logger
	srv    *http.Server
}

// Logger is the subset of slog.Logger we need.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// New constructs a Server.
func New(cfg *config.Config, logger Logger) *Server {
	return &Server{cfg: cfg, logger: logger}
}

// Start runs the HTTP server until context is canceled.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/projects", s.handleProjects)
	mux.HandleFunc("/api/projects/", s.handleProjectScoped)

	// Static UI.
	sub, _ := fs.Sub(embeddedFS, "web")
	mux.Handle("/", http.FileServer(http.FS(sub)))

	s.srv = &http.Server{
		Addr:              s.cfg.UI.Addr,
		Handler:           s.withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("ui server listening", "addr", s.cfg.UI.Addr)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// Middleware to allow local JS fetches.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if tok := strings.TrimSpace(s.cfg.UI.AuthToken); tok != "" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer "+tok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.writeJSON(w, s.cfg.Projects)
}

func (s *Server) handleProjectScoped(w http.ResponseWriter, r *http.Request) {
	// Expect /api/projects/{id}/issues or /api/projects/{id}/issues/{issueId}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	projectID := parts[0]
	project, ok := s.findProject(projectID)
	if !ok {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if parts[1] != "issues" {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 2 {
		// /issues collection
		switch r.Method {
		case http.MethodGet:
			s.listIssues(w, r, project)
		case http.MethodPost:
			s.createIssue(w, r, project)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /issues/{id}
	issueID := parts[2]
	switch r.Method {
	case http.MethodGet:
		s.getIssue(w, r, project, issueID)
	case http.MethodPatch:
		s.updateIssue(w, r, project, issueID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listIssues(w http.ResponseWriter, r *http.Request, p config.Project) {
	q := r.URL.Query()
	status := q.Get("status")
	issueType := q.Get("type")
	limit := q.Get("limit")
	if limit == "" {
		limit = "50"
	}

	args := []string{"list", "--json", "-n", limit}
	if status != "" {
		args = append(args, "--status", status)
	}
	if issueType != "" {
		args = append(args, "--type", issueType)
	}
	if labelAny := q.Get("label_any"); labelAny != "" {
		args = append(args, "--label-any", labelAny)
	}

	if labelAny := q.Get("label_any"); labelAny != "" {
		args = append(args, "--label-any", labelAny)
	}

	out, err := runBd(r.Context(), p.Path, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func (s *Server) getIssue(w http.ResponseWriter, r *http.Request, p config.Project, issueID string) {
	out, err := runBd(r.Context(), p.Path, "list", "--json", "--id", issueID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func (s *Server) createIssue(w http.ResponseWriter, r *http.Request, p config.Project) {
	var body struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Type        string   `json:"type"` // epic|feature|task|bug|chore
		Parent      string   `json:"parent"`
		Labels      []string `json:"labels"`
		Priority    string   `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	issueType := body.Type
	if issueType == "" {
		issueType = "task"
	}

	args := []string{"create", "--json", "--type", issueType, "--title", body.Title}
	if body.Description != "" {
		args = append(args, "--description", body.Description)
	}
	if body.Parent != "" {
		args = append(args, "--parent", body.Parent)
	}
	if body.Priority != "" {
		args = append(args, "--priority", body.Priority)
	}
	for _, l := range body.Labels {
		args = append(args, "--labels", l)
	}

	out, err := runBd(r.Context(), p.Path, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func (s *Server) updateIssue(w http.ResponseWriter, r *http.Request, p config.Project, issueID string) {
	var body struct {
		Title        *string  `json:"title"`
		Description  *string  `json:"description"`
		Status       *string  `json:"status"`
		Priority     *string  `json:"priority"`
		AddLabels    []string `json:"addLabels"`
		RemoveLabels []string `json:"removeLabels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	args := []string{"update", "--json", issueID}
	if body.Title != nil {
		args = append(args, "--title", *body.Title)
	}
	if body.Description != nil {
		args = append(args, "--description", *body.Description)
	}
	if body.Status != nil {
		args = append(args, "--status", *body.Status)
	}
	if body.Priority != nil {
		args = append(args, "--priority", *body.Priority)
	}
	for _, l := range body.AddLabels {
		args = append(args, "--add-label", l)
	}
	for _, l := range body.RemoveLabels {
		args = append(args, "--remove-label", l)
	}

	out, err := runBd(r.Context(), p.Path, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func (s *Server) findProject(id string) (config.Project, bool) {
	for _, p := range s.cfg.Projects {
		if p.ID == id {
			return p, true
		}
	}
	return config.Project{}, false
}

func (s *Server) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// runBd runs the bd CLI in the given directory.
func runBd(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "bd", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bd %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return out, nil
}

// sanitizePath cleans the provided path (not heavily used yet).
func sanitizePath(p string) string {
	if p == "" {
		return p
	}
	return filepath.Clean(path.Clean(p))
}
