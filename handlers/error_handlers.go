package handlers

import (
	"fmt"
	"log"
	"net/http"
)

// ErrorHandler renders a response when an error occurs
func TemplateErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Template error handler returning 500")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(500)
	fmt.Fprintln(w, err.Error())
}

func KratosErrorHandler(w http.ResponseWriter, r *http.Request, response *http.Response, err error, redirect string) {
	log.Printf("Kratos error handler redirecting to %v", redirect)
	if response.StatusCode == 404 || response.StatusCode == 410 || response.StatusCode == 403 {
		http.Redirect(w, r, redirect, http.StatusMovedPermanently)
		return
	}
}
