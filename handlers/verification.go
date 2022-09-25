package handlers

import (
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
)

// VerificationParams configure the Login http handler
type VerificationParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// FlowRedirectURL is the kratos URL to redirect the browser to,
	// when the user wishes to login, and the 'flow' query param is missing
	FlowRedirectURL string
}

// Login handler displays the login screen
func (vp VerificationParams) Verification(w http.ResponseWriter, r *http.Request) {

	// Start the verification flow with Kratos if required
	flow := r.URL.Query().Get("flow")
	if flow == "" {
		log.Printf("No flow ID found in URL, initializing login flow, redirect to %s", vp.FlowRedirectURL)
		http.Redirect(w, r, vp.FlowRedirectURL, http.StatusMovedPermanently)
		return
	}

	log.Print("Calling Kratos API to get self service verification")
	verificationResp, rawResp, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceVerificationFlow(r.Context()).Id(flow).Cookie(r.Header.Get("Cookie")).Execute()
	if err != nil {
		KratosErrorHandler(w, r, rawResp, err, vp.FlowRedirectURL)
		return
	}

	dataMap := map[string]interface{}{
		"title": "Verify account",
		"resp":  verificationResp,
		"fs":    vp.FS,
	}
	if err = GetTemplate(verificationPage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
