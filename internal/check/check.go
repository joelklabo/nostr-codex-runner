package check

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Result represents a single dependency check outcome.
type Result struct {
	Name     string
	Type     string
	Status   string // OK|MISSING|WARN
	Details  string
	Optional bool
}

// Checker defines an interface for running checks.
type Checker interface {
	Check(dep DepInput) Result
}

// DepInput is a simplified view from config.Dep.
type DepInput struct {
	Name        string
	Type        string
	Version     string
	Optional    bool
	Description string
	Hint        string
}

// BinaryChecker checks for a binary on PATH and optional version substring.
type BinaryChecker struct{}

func (BinaryChecker) Check(dep DepInput) Result {
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	path, err := exec.LookPath(dep.Name)
	if err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("not found in PATH (%s)", dep.Hint)
		return res
	}
	if dep.Version != "" {
		out, _ := exec.Command(path, "--version").CombinedOutput()
		if !strings.Contains(string(out), dep.Version) {
			res.Status = missingStatus(dep.Optional)
			res.Details = fmt.Sprintf("found %s but version mismatch (need %s)", strings.TrimSpace(string(out)), dep.Version)
			return res
		}
	}
	res.Details = path
	return res
}

// EnvChecker ensures an environment variable is present.
type EnvChecker struct{}

func (EnvChecker) Check(dep DepInput) Result {
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	if v, ok := os.LookupEnv(dep.Name); !ok || v == "" {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("env var %s not set (%s)", dep.Name, dep.Hint)
		return res
	}
	return res
}

// FileChecker ensures a file or directory exists at the given path.
type FileChecker struct{}

func (FileChecker) Check(dep DepInput) Result {
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	if dep.Name == "" {
		res.Status = missingStatus(dep.Optional)
		res.Details = "path not provided"
		return res
	}
	if _, err := os.Stat(dep.Name); err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("%s (%s)", err.Error(), dep.Hint)
		return res
	}
	return res
}

// URLChecker performs a simple GET with a short timeout.
type URLChecker struct{}

func (URLChecker) Check(dep DepInput) Result {
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(dep.Name)
	if err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("request failed (%s)", dep.Hint)
		return res
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("status %d (%s)", resp.StatusCode, dep.Hint)
		return res
	}
	return res
}

// PortChecker tries to connect to host:port.
type PortChecker struct{}

func (PortChecker) Check(dep DepInput) Result {
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	conn, err := net.DialTimeout("tcp", dep.Name, 2*time.Second)
	if err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("unreachable (%s)", dep.Hint)
		return res
	}
	_ = conn.Close()
	return res
}

// RelayChecker tries to reach a relay host:port (wss/ws/http style).
type RelayChecker struct{}

func (RelayChecker) Check(dep DepInput) Result {
	target, err := relayTarget(dep.Name)
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	if err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("invalid relay: %v", err)
		return res
	}
	conn, err := net.DialTimeout("tcp", target, 2*time.Second)
	if err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("unreachable (%s)", dep.Hint)
		return res
	}
	_ = conn.Close()
	return res
}

// DirWriteChecker ensures a directory is writable by creating a temp file.
type DirWriteChecker struct{}

func (DirWriteChecker) Check(dep DepInput) Result {
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	if dep.Name == "" {
		res.Status = missingStatus(dep.Optional)
		res.Details = "directory not provided"
		return res
	}
	f, err := os.CreateTemp(dep.Name, ".buddy-check-*")
	if err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("not writable (%s)", dep.Hint)
		return res
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return res
}

func relayTarget(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("empty relay")
	}
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", err
		}
		host := u.Host
		if !strings.Contains(host, ":") {
			port := "443"
			if u.Scheme == "ws" || u.Scheme == "http" {
				port = "80"
			}
			host = host + ":" + port
		}
		return host, nil
	}
	// assume host:port already
	return raw, nil
}

func missingStatus(optional bool) string {
	if optional {
		return "WARN"
	}
	return "MISSING"
}
