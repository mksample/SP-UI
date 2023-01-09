package middleware

import (
	"log"
	"net/http"

	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
	"github.com/davidoram/kratos-selfservice-ui-go/session"
)

const code2FA = 403

// KratosAuthParams configure the KratosAuth http handler
type KratosAuthParams struct {
	session.SessionStore

	// RedirectUnauthURL is where we will rerirect to if the session is
	// not associated with a valid user
	RedirectUnauthURL string

	// Redirect2FA is where we will redirect to if the session requires 2FA authenication.
	Redirect2FA string
}

// KratoAuthMiddleware retrieves the user from the session via Kratos WhoAmIURL,
// and if the user is authenticated the request will proceed through the middleware chain.
// If the session is not authenticated, redirects to the RedirectUnauthURL
func (p KratosAuthParams) KratoAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, rawResp, err := api_client.PublicClient().V0alpha2Api.ToSession(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
		if rawResp != nil && rawResp.StatusCode == code2FA {
			log.Printf("2 factor authentication required, redirecting to %v", p.Redirect2FA)
			http.Redirect(w, r, p.Redirect2FA, http.StatusPermanentRedirect)
		} else if err != nil {
			log.Printf("No kratos session found: %v, redirecting to %v", err, p.RedirectUnauthURL)
			http.Redirect(w, r, p.RedirectUnauthURL, http.StatusPermanentRedirect)
		} else {
			err = p.SaveKratosSession(w, r, session)
			if err != nil {
				log.Printf("Error saving kratos session: %v", err)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SetSession attempts to set the session for the request. If the user is not authenicated no session is set.
// Redirects to MFA login if required.
func (p KratosAuthParams) SetSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, rawResp, err := api_client.PublicClient().V0alpha2Api.ToSession(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
		if rawResp != nil && rawResp.StatusCode == code2FA {
			log.Printf("2 factor authentication required, redirecting to %v", p.Redirect2FA)
			http.Redirect(w, r, p.Redirect2FA, http.StatusPermanentRedirect)
		} else if rawResp != nil && rawResp.StatusCode == 401 {
			log.Printf("No session to set")
		} else if err != nil {
			log.Printf("Error setting kratos session: %v", err)
		} else {
			err = p.SaveKratosSession(w, r, session)
			if err != nil {
				log.Printf("Error saving kratos session: %v", err)
			}
		}

		next.ServeHTTP(w, r)
	})
}
