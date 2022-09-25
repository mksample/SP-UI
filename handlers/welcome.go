package handlers

import (
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
	var logoutURL string
	logoutResp, rawResp, err := api_client.PublicClient().V0alpha2Api.CreateSelfServiceLogoutFlowUrlForBrowsers(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
	if rawResp != nil && rawResp.StatusCode == 401 {
		logoutURL = ""
	} else if err != nil {
		log.Printf("Getting logout url: %v", err)
	} else {
		logoutURL = logoutResp.GetLogoutUrl()
	}

	log.Printf("Getting session from session store")
	sessionStr := `No valid Ory Session was found.
		Please sign in to receive one.`
	if wp.HasKratosSession(r) {
		byteSessionStr, err := wp.GetKratosSession(r).MarshalJSON()
		if err != nil {
			log.Printf("Error marshaling session to json: %v", err)
		} else {
			sessionStr = string(byteSessionStr)
		}
	}

	dataMap := map[string]interface{}{
		"title":      "Welcome to Ory",
		"session":    sessionStr,
		"hasSession": wp.HasKratosSession(r),
		"logoutUrl":  logoutURL,
		"fs":         wp.FS,
	}
	if err := GetTemplate(welcomePage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
