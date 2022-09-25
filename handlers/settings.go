package handlers

import (
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
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
	settingsResp, rawResp, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceSettingsFlow(r.Context()).Id(flow).Cookie(r.Header.Get("Cookie")).Execute()
	if err != nil {
		KratosErrorHandler(w, r, rawResp, err, sp.FlowRedirectURL)
		return
	}

	dataMap := map[string]interface{}{
		"title": "Account settings",
		"resp":  settingsResp,
		"fs":    sp.FS,
	}
	if err = GetTemplate(settingsPage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
