package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
	"github.com/davidoram/kratos-selfservice-ui-go/session"
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
	logout, _, err := api_client.PublicClient().V0alpha2Api.CreateSelfServiceLogoutFlowUrlForBrowsers(context.Background()).Cookie(session.SessionCookieName).Execute()
	if err != nil {
		log.Printf("Error creating self service logout flow: %v, redirecting to /", err)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	log.Print("Calling Kratos API to get self service login")
	loginResp, _, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceLoginFlow(context.Background()).Id(flow).Cookie(session.SessionCookieName).Execute()
	if err != nil {
		log.Printf("Error getting self service login flow: %v, redirecting to /", err)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	dataMap := map[string]interface{}{
		"resp":            loginResp,
		"isAuthenticated": *loginResp.Refresh || loginResp.RequestedAal == kratos.AUTHENTICATORASSURANCELEVEL_AAL2.Ptr(),
		"signUpUrl":       lp.RegistrationURL,
		"logoutUrl":       logout.LogoutUrl,
		"fs":              lp.FS,
	}
	if err = GetTemplate(loginPage).Render("layout", w, r, dataMap); err != nil {
		ErrorHandler(w, r, err)
	}
}
