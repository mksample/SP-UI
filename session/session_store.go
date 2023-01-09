// session package provides typesafe access to session data
package session

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	client "github.com/ory/kratos-client-go"
)

// SessionStore holds a connection to the application Session store
type SessionStore struct {
	// Session store
	Store *sessions.CookieStore
}

const (
	// The cookie we use to store this applications session in
	SessionCookieName = "kgc-sess"

	// Keys we store in our application session
	keyKratosSession = "kratosSession"
)

// SaveKratosSession stores a kratos session in the session store.
func (s SessionStore) SaveKratosSession(w http.ResponseWriter, r *http.Request, ks *client.Session) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := s.Store.Get(r, SessionCookieName)
	if err != nil {
		log.Printf("Error decoding session, %v", err)
		return err
	}

	// Add the value into the session store and set the expiry
	session.Values[keyKratosSession] = *ks
	session.Options.MaxAge = int(ks.ExpiresAt.Unix()) - int(time.Now().Unix())

	// Save it before we write to the response/return from the handler.
	return session.Save(r, w)
}

// GetKratosSession returns a krato session from the session store.
func (s SessionStore) GetKratosSession(r *http.Request) *client.Session {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := s.Store.Get(r, SessionCookieName)
	if err != nil {
		log.Printf("Error decoding session, %v", err)
		return nil
	}
	if v, exists := session.Values[keyKratosSession]; exists {
		ks := v.(client.Session)
		return &ks
	}
	return nil
}

// HasKratosSession checks if there is a kratos session in the session store.
func (s SessionStore) HasKratosSession(r *http.Request) bool {
	session, err := s.Store.Get(r, SessionCookieName)
	if err != nil {
		log.Printf("Error decoding session, %v", err)
		return false
	}
	if _, exists := session.Values[keyKratosSession]; exists {
		return true
	}
	return false
}

// ClearKratosSession clears any existing kratos session in the session store.
func (s SessionStore) ClearKratosSession(w http.ResponseWriter, r *http.Request) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := s.Store.Get(r, SessionCookieName)
	if err != nil {
		log.Printf("Error decoding session, %v", err)
		return err
	}

	// Clear the value stored in the session store
	delete(session.Values, keyKratosSession)
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
