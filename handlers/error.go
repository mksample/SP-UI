package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
)

// ErrorParams configure the Login http handler
type ErrorParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// RedirectURL is the kratos URL to redirect the browser to if an error occurs
	RedirectURL string
}

// Login handler displays the login screen
func (ep ErrorParams) Error(w http.ResponseWriter, r *http.Request) {

	// Start the error flow with Kratos if required
	flow := r.URL.Query().Get("flow")
	if flow == "" {
		log.Printf("No error was sent, redirecting to %s", ep.RedirectURL)
		http.Redirect(w, r, ep.RedirectURL, http.StatusMovedPermanently)
		return
	}

	log.Print("Calling Kratos API to get self service error")
	errorResp, _, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceError(context.Background()).Id(flow).Execute()
	if err != nil {
		log.Printf("Error getting self service error flow: %v, redirecting to %s", err, ep.RedirectURL)
		http.Redirect(w, r, ep.RedirectURL, http.StatusMovedPermanently)
		return
	}

	dataMap := map[string]interface{}{
		"message": errorResp.GetError(),
		"fs":      ep.FS,
	}
	if err = GetTemplate(errorPage).Render("layout", w, r, dataMap); err != nil {
		ErrorHandler(w, r, err)
	}
}
