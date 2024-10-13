package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

type LoggingMiddleware struct {
	logger Logger
	next   http.Handler
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rw := &responseWriter{w, http.StatusOK}
	l.next.ServeHTTP(rw, r)
	latency := time.Since(start)
	msg := fmt.Sprintf(
		"[http] %s %s %d %s (%s, User agent: %s, IP: %s)",
		r.Method, r.URL.Path, rw.status, latency, r.Proto, r.UserAgent(), r.RemoteAddr,
	)
	l.logger.Info(msg)
}
