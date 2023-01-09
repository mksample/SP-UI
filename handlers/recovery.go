package handlers

import (
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
)

// RecoveryParams configure the Login http handler
type RecoveryParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// FlowRedirectURL is the kratos URL to redirect the browser to,
	// when the user wishes to login, and the 'flow' query param is missing
	FlowRedirectURL string
}

// Login handler displays the login screen
func (rp RecoveryParams) Recovery(w http.ResponseWriter, r *http.Request) {

	// Start the recovery flow with Kratos if required
	flow := r.URL.Query().Get("flow")
	if flow == "" {
		http.Redirect(w, r, rp.FlowRedirectURL, http.StatusMovedPermanently)
		return
	}

	recoveryResp, rawResp, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceRecoveryFlow(r.Context()).Id(flow).Cookie(r.Header.Get("Cookie")).Execute()
	if err != nil {
		KratosErrorHandler(w, r, rawResp, err, rp.FlowRedirectURL)
		return
	}

	dataMap := map[string]interface{}{
		"title": "Recover account",
		"resp":  recoveryResp,
		"fs":    rp.FS,
	}
	if err = GetTemplate(recoveryPage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
