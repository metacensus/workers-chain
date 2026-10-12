package healthcheck

import (
	"context"
	"net/http"

	"go.temporal.io/sdk/client"
)

const (
	HealthCheckRoute = "/health"
	ReadyCheckRoute  = "/ready"
)

type Server struct {
	temporalClient client.Client
}

func NewServer(c client.Client) *Server {
	return &Server{
		temporalClient: c,
	}
}

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	// Check Temporal connection
	ctx := context.Background()
	_, err := s.temporalClient.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc(HealthCheckRoute, s.Handler)
	mux.HandleFunc(ReadyCheckRoute, s.Handler)
	return http.ListenAndServe(addr, mux)
}
