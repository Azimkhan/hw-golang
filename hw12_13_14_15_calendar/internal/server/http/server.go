package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Server struct {
	logger     Logger
	app        Application
	mux        *http.ServeMux
	httpServer *http.Server
	bindAddr   string
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Application interface {
}

func NewServer(logger Logger, gRPCHandler http.HandlerFunc, app Application, bindAddr string) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", gRPCHandler)
	mux.HandleFunc("/hello", HelloHandler)
	return &Server{
		logger:   logger,
		app:      app,
		mux:      mux,
		bindAddr: bindAddr,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:              s.bindAddr,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           &LoggingMiddleware{logger: s.logger, next: s.mux},
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	s.logger.Info(fmt.Sprintf("http server is running on %s", s.bindAddr))
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.httpServer.Shutdown(ctx)
	return err
}
