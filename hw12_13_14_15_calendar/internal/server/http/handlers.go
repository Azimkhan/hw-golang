package internalhttp

import "net/http"

func HelloHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello!"))
}
