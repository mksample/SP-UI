package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
	"github.com/davidoram/kratos-selfservice-ui-go/session"
)

// WelcomeParams configure the Login http handler
type WelcomeParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// FlowRedirectURL is the kratos URL to redirect the browser to,
	// when the user wishes to login, and the 'flow' query param is missing
	FlowRedirectURL string
	session.SessionStore
}

// Login handler displays the login screen
func (wp WelcomeParams) Welcome(w http.ResponseWriter, r *http.Request) {

	log.Printf("Calling Kratos API to create self service logout")
	logoutResp, _, err := api_client.PublicClient().V0alpha2Api.CreateSelfServiceLogoutFlowUrlForBrowsers(context.Background()).Cookie(session.SessionCookieName).Execute()
	if err != nil {
		log.Printf("Error creating self service logout flow: %v, redirecting to /", err)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	sessionStr := wp.GetKratosSession(r).JsonPretty()
	if sessionStr == "" {
		sessionStr = `No valid Ory Session was found.
		Please sign in to receive one.`
	}

	dataMap := map[string]interface{}{
		"session":    sessionStr,
		"hasSession": wp.HasKratosSession(r),
		"logoutUrl":  logoutResp.LogoutUrl,
		"fs":         wp.FS,
	}
	if err = GetTemplate(welcomePage).Render("layout", w, r, dataMap); err != nil {
		ErrorHandler(w, r, err)
	}
}
