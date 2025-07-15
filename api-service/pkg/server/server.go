package server

import (
	"net/http"
	"time"
)

type ServerHTTP struct {
	httpServer *http.Server
}

func NewServerHTTPClient(port string, handler http.Handler) *ServerHTTP {
	return &ServerHTTP{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  5 * time.Second,   // max time to read request from the client
			WriteTimeout: 10 * time.Second,  // max time to write response to the client
			IdleTimeout:  120 * time.Second, // max time for connections using TCP
		},
	}
}
func (s *ServerHTTP) Run() error {
	return s.httpServer.ListenAndServe()
}
