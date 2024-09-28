package internalhttp

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Server struct {
	logger     Logger
	app        Application
	mux        *http.ServeMux
	httpServer *http.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Application interface {
	GetHTTPBindAddr() string
}

func NewServer(logger Logger, app Application) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", HelloHandler)
	return &Server{
		logger: logger,
		app:    app,
		mux:    mux,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:              ":8081",
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           &LoggingMiddleware{logger: s.logger, next: s.mux},
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.httpServer.Shutdown(ctx)
	return err
}

// TODO
