package handlers

import (
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
)

// RegistrationParams configure the Login http handler
type RegistrationParams struct {
	// FS provides access to static files
	FS *hashfs.FS

	// FlowRedirectURL is the kratos URL to redirect the browser to,
	// when the user wishes to login, and the 'flow' query param is missing
	FlowRedirectURL string
	LoginURL        string
}

// Login handler displays the login screen
func (rp RegistrationParams) Registration(w http.ResponseWriter, r *http.Request) {
	// Start the registration flow with Kratos if required
	flow := r.URL.Query().Get("flow")
	if flow == "" {
		http.Redirect(w, r, rp.FlowRedirectURL, http.StatusMovedPermanently)
		return
	}

	registrationResp, rawResp, err := api_client.PublicClient().V0alpha2Api.GetSelfServiceRegistrationFlow(r.Context()).Id(flow).Cookie(r.Header.Get("Cookie")).Execute()
	if err != nil {
		KratosErrorHandler(w, r, rawResp, err, rp.FlowRedirectURL)
		return
	}

	dataMap := map[string]interface{}{
		"title":     "Create account",
		"resp":      registrationResp,
		"signInUrl": rp.LoginURL,
		"fs":        rp.FS,
	}
	if err = GetTemplate(registrationPage).Render("layout", w, r, dataMap); err != nil {
		TemplateErrorHandler(w, r, err)
	}
}
