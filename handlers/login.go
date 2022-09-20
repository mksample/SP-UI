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

	log.Printf("Calling Kratos API to create self service logout")
	var logoutURL string
	logoutResp, rawResp, err := api_client.PublicClient().V0alpha2Api.CreateSelfServiceLogoutFlowUrlForBrowsers(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
	if rawResp != nil && rawResp.StatusCode == 401 {
		logoutURL = ""
	} else if rawResp == nil && err != nil {
		log.Printf("Getting logout url: %v", err)
	} else {
		logoutURL = logoutResp.GetLogoutUrl()
	}

	log.Print("Calling Kratos API to get self service login")
	loginResp, _, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceLoginFlow(r.Context()).Id(flow).Cookie(r.Header.Get("Cookie")).Execute()
	if err != nil {
		log.Printf("Error getting self service login flow: %v, redirecting to /", err)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	dataMap := map[string]interface{}{
		"resp":            loginResp,
		"isAuthenticated": *loginResp.Refresh || loginResp.RequestedAal == kratos.AUTHENTICATORASSURANCELEVEL_AAL2.Ptr(),
		"signUpUrl":       lp.RegistrationURL,
		"logoutUrl":       logoutURL,
		"fs":              lp.FS,
	}
	if err = GetTemplate(loginPage).Render("layout", w, r, dataMap); err != nil {
		ErrorHandler(w, r, err)
	}
}
