package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
	"github.com/davidoram/kratos-selfservice-ui-go/session"
)

// SettingsParams configure the Login http handler
type SettingsParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// FlowRedirectURL is the kratos URL to redirect the browser to,
	// when the user wishes to login, and the 'flow' query param is missing
	FlowRedirectURL string
}

// Login handler displays the login screen
func (sp SettingsParams) Settings(w http.ResponseWriter, r *http.Request) {

	// Start the settings flow with Kratos if required
	flow := r.URL.Query().Get("flow")
	if flow == "" {
		log.Printf("No flow ID found in URL, initializing login flow, redirect to %s", sp.FlowRedirectURL)
		http.Redirect(w, r, sp.FlowRedirectURL, http.StatusMovedPermanently)
		return
	}

	log.Print("Calling Kratos API to get self service settings")
	settingsResp, _, err := api_client.AdminClient().V0alpha2Api.GetSelfServiceSettingsFlow(context.Background()).Id(flow).Cookie(session.SessionCookieName).Execute()
	if err != nil {
		log.Printf("Error getting self service settings flow: %v, redirecting to /", err)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	dataMap := map[string]interface{}{
		"resp": settingsResp,
		"fs":   sp.FS,
	}
	if err = GetTemplate(settingsPage).Render("layout", w, r, dataMap); err != nil {
		ErrorHandler(w, r, err)
	}
}
