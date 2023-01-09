package handlers

import (
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
)

// ErrorParams configure the Login http handler
type KratosErrorParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// RedirectURL is the URL to redirect the browser to if an error occurs
	RedirectURL string
	// HomeURL is the URL for returning home
	HomeURL string
}

// Login handler displays the login screen
func (ep KratosErrorParams) Error(w http.ResponseWriter, r *http.Request) {

	// Start the error flow with Kratos if required
	flow := r.URL.Query().Get("flow")
	if flow == "" {
		http.Redirect(w, r, ep.RedirectURL, http.StatusMovedPermanently)
		return
	}

	errorResp, rawResp, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceError(r.Context()).Id(flow).Execute()
	if err != nil {
		log.Printf("Error getting self service error flow: %v, redirecting to %s", err, ep.RedirectURL)
		http.Redirect(w, r, ep.RedirectURL, http.StatusMovedPermanently)
		return
	} else if rawResp != nil {
		if rawResp.StatusCode == 404 {
			log.Printf("Error could not be found, redirecting to %v", ep.RedirectURL)
			http.Redirect(w, r, ep.RedirectURL, http.StatusMovedPermanently)
		}
	}

	dataMap := map[string]interface{}{
		"title":   "An error occurred",
		"homeURL": ep.HomeURL,
		"message": errorResp.GetError(),
		"fs":      ep.FS,
	}
	if err = GetTemplate(errorPage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
