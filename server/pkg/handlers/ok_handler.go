package handlers

import (
	"fmt"
	"net/http"
)

// OKHandler is a simple handler that returns a 200 OK status code
func OKHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}