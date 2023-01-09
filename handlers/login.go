package handlers

import (
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
	kratos "github.com/ory/kratos-client-go"
)

// LoginParams configure the Login http handler
type LoginParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// FlowRedirectURL is the kratos URL to redirect the browser to,
	// when the user wishes to login, and the 'flow' query param is missing
	FlowRedirectURL string
	RegistrationURL string
}

// Login handler displays the login screen
func (lp LoginParams) Login(w http.ResponseWriter, r *http.Request) {

	// Start the login flow with Kratos if required
	flow := r.URL.Query().Get("flow")
	if flow == "" {
		log.Printf("No flow ID found in URL, initializing login flow, redirect to %s", lp.FlowRedirectURL)
		http.Redirect(w, r, lp.FlowRedirectURL, http.StatusMovedPermanently)
		return
	}

	var logoutURL string
	logoutResp, rawResp, err := api_client.PublicClient().V0alpha2Api.CreateSelfServiceLogoutFlowUrlForBrowsers(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
	if rawResp != nil && rawResp.StatusCode == 401 {
		logoutURL = ""
	} else if err != nil {
		log.Printf("Error getting logout url: %v", err)
	} else {
		logoutURL = logoutResp.GetLogoutUrl()
	}

	loginResp, rawResp, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceLoginFlow(r.Context()).Id(flow).Cookie(r.Header.Get("Cookie")).Execute()
	if err != nil {
		KratosErrorHandler(w, r, rawResp, err, lp.FlowRedirectURL)
		return
	}

	dataMap := map[string]interface{}{
		"title":           "Sign in",
		"resp":            loginResp,
		"isAuthenticated": *loginResp.Refresh || loginResp.RequestedAal == kratos.AUTHENTICATORASSURANCELEVEL_AAL2.Ptr(),
		"registrationURL": lp.RegistrationURL,
		"logoutURL":       logoutURL,
		"fs":              lp.FS,
	}
	if err = GetTemplate(loginPage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
