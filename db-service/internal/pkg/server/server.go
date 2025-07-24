package server

import (
	"context"
	"net/http"
	"time"
)

type ServerHTTP struct {
	httpServer *http.Server
}

func NewServerHTTPClient(port string, handler http.Handler) *ServerHTTP {
	return &ServerHTTP{
		httpServer: &http.Server{
			Addr:           ":" + port,
			Handler:        handler,
			MaxHeaderBytes: 1 << 20, // 1 MB
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
		},
	}
}

func (s *ServerHTTP) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *ServerHTTP) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
