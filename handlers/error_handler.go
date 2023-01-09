package handlers

import (
	"log"
	"net/http"
)

var ErrorURL string

func InitErrorHandler(URL string) {
	ErrorURL = URL
}

// ErrorHandler renders a response when an error occurs
func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("ErrorHandler returning 500")
	w.WriteHeader(500)
	http.Redirect(w, r, ErrorURL, http.StatusMovedPermanently)
}
