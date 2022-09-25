package handlers

import (
	"net/http"

	"github.com/benbjohnson/hashfs"
)

// ErrorParams configure the Login http handler
type PageNotFoundParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// HomeURL is the URL for returning home
	HomeURL string
}

// Login handler displays the login screen
func (pp PageNotFoundParams) PageNotFound(w http.ResponseWriter, r *http.Request) {
	dataMap := map[string]interface{}{
		"homeURL": pp.HomeURL,
		"message": "The requested page could not be found (404)",
		"fs":      pp.FS,
	}
	if err := GetTemplate(errorPage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
