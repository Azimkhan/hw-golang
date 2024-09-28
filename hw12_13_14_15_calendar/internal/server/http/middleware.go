package internalhttp

import (
	"net/http"
)

type LoggingMiddleware struct {
	logger Logger
	next   http.Handler
}

func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.logger.Info("request")
	l.next.ServeHTTP(w, r)
}
