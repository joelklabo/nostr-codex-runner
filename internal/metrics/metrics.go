package metrics

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	inboundMsgs = prometheus.NewCounter(prometheus.CounterOpts{Name: "runner_inbound_total", Help: "Inbound messages seen"})
	agentErrors = prometheus.NewCounter(prometheus.CounterOpts{Name: "runner_agent_errors_total", Help: "Agent errors"})
	actionCalls = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "runner_action_calls_total", Help: "Action invocations"}, []string{"action", "status"})
	sendErrors  = prometheus.NewCounter(prometheus.CounterOpts{Name: "runner_send_errors_total", Help: "Transport send errors"})
)

func init() {
	prometheus.MustRegister(inboundMsgs, agentErrors, actionCalls, sendErrors)
}

// Start runs a Prometheus handler on the given listen addr.
func Start(ctx context.Context, listen string, log *slog.Logger) error {
	if listen == "" {
		return nil
	}
	srv := &http.Server{Handler: promhttp.Handler()}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if log != nil {
				log.Error("metrics server failed", slog.String("err", err.Error()))
			}
		}
	}()
	return nil
}

func IncInbound() { inboundMsgs.Inc() }

func IncAgentError() { agentErrors.Inc() }

func IncAction(action string, status string) { actionCalls.WithLabelValues(action, status).Inc() }

func IncSendError() { sendErrors.Inc() }
