package main

import (
	"context"
	"embed"
	"encoding/gob"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/davidoram/kratos-selfservice-ui-go/api_client"
	"github.com/davidoram/kratos-selfservice-ui-go/handlers"
	"github.com/davidoram/kratos-selfservice-ui-go/middleware"
	"github.com/davidoram/kratos-selfservice-ui-go/options"
	"github.com/davidoram/kratos-selfservice-ui-go/session"

	"github.com/benbjohnson/hashfs"
	kratos "github.com/ory/kratos-client-go"

	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// staticFS holds the static files, CSS images etc.
// Its baked into the application executable using the embed API - see https://golang.org/pkg/embed/
//go:embed static
var staticFS embed.FS

func main() {
	opt := options.NewOptions().SetFromCommandLine()
	if err := opt.Validate(); err != nil {
		log.Fatalf("Error parsing command line: %v", err)
	}
	log.Printf("KratosAdminURL: %s", opt.KratosAdminURL.String())
	log.Printf("KratosPublicURL: %s", opt.KratosPublicURL.String())
	log.Printf("KratosBrowserURL: %s", opt.KratosBrowserURL.String())
	log.Printf("BaseURL: %s", opt.BaseURL.String())
	log.Printf("Address: %s", opt.Address())
	log.Printf("Port: %v", opt.Port)
	log.Printf("Number of Cookie store keys: %d", len(opt.CookieStoreKeyPairs))

	// Init API clients
	if _, err := api_client.InitPublicClient(opt); err != nil {
		log.Fatalf("Error initializing public API client failed with error: %v", err)
	}
	if _, err := api_client.InitAdminClient(opt); err != nil {
		log.Fatalf("Error initializing admin API client failed with error: %v", err)
	}

	// Setup sesssion store in cookies
	var store = sessions.NewCookieStore(opt.CookieStoreKeyPairs...)

	// Register kratos session type with gob
	gob.Register(kratos.Session{})
	gob.Register(make(map[string]interface{}))

	// Create router
	r := mux.NewRouter()

	// Static assets are wrapped in a hash fs that allows for aggesive http caching
	var fsys = hashfs.NewFS(staticFS)
	r.PathPrefix("/static/").Handler(hashfs.FileServer(fsys))

	// Public Routes
	r.Use(gh.RecoveryHandler(gh.PrintRecoveryStack(true)), middleware.NoCacheMiddleware)

	// Health/readiness probe endpoints
	r.HandleFunc("/health/alive", handlers.Health)
	r.HandleFunc("/health/ready", handlers.Health)

	// Redirect from / to /welcome
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/welcome", http.StatusMovedPermanently)
	})

	// Login page
	loginP := handlers.LoginParams{
		FlowRedirectURL: opt.LoginFlowURL(),
		RegistrationURL: opt.RegistrationURL(),
		FS:              fsys,
	}
	r.HandleFunc("/login", loginP.Login).Name("login")

	// Registration page
	regP := handlers.RegistrationParams{
		FlowRedirectURL: opt.RegistrationURL(),
		LoginURL:        opt.LoginURL(),
		FS:              fsys,
	}
	r.HandleFunc("/registration", regP.Registration)

	// Verification page
	verificationP := handlers.VerificationParams{
		FlowRedirectURL: opt.VerificationURL(),
		FS:              fsys,
	}
	r.HandleFunc("/verification", verificationP.Verification)

	// Recovery page
	recoverP := handlers.RecoveryParams{
		FlowRedirectURL: opt.RecoveryFlowURL(),
		FS:              fsys,
	}
	r.HandleFunc("/recovery", recoverP.Recovery)

	// Error page
	errorP := handlers.KratosErrorParams{
		RedirectURL: opt.GetBaseURL(),
		HomeURL:     opt.GetBaseURL(),
		FS:          fsys,
	}
	r.HandleFunc("/error", errorP.Error)

	// 404 route
	pageNotFoundP := handlers.PageNotFoundParams{
		HomeURL: opt.GetBaseURL(),
		FS:      fsys,
	}
	r.NotFoundHandler = http.HandlerFunc(pageNotFoundP.PageNotFound)

	// Routes with authentication middleware
	authP := middleware.KratosAuthParams{
		SessionStore:      session.SessionStore{Store: store},
		RedirectUnauthURL: MustURL(r.Get("login")).String(),
		Redirect2FA:       opt.TwoFAURL(),
	}

	// Welcome page (authentication optional)
	welcomeP := handlers.WelcomeParams{
		SessionStore: session.SessionStore{Store: store},
		FS:           fsys,
	}
	r.Handle("/welcome", Middleware(
		http.HandlerFunc(welcomeP.Welcome),
		authP.SetSession,
	))

	// Settings page (authentication required)
	settingsP := handlers.SettingsParams{
		FlowRedirectURL: opt.SettingsURL(),
		FS:              fsys,
	}
	r.Handle("/settings", Middleware(
		http.HandlerFunc(settingsP.Settings),
		authP.KratoAuthMiddleware,
	))

	// Wrap everything in a logger
	logR := gh.LoggingHandler(os.Stdout, r)

	// Start server
	srv := &http.Server{
		Addr: opt.Address(),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      logR, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if opt.TLSCertPath != "" {
			log.Printf("Serving TLS")
			if err := srv.ListenAndServeTLS(opt.TLSCertPath, opt.TLSKeyPath); err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("Serving")
			if err := srv.ListenAndServe(); err != nil {
				log.Println(err)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), opt.ShutdownWait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

// MustURL returns a 'named' URL or panics
func MustURL(r *mux.Route, pairs ...string) *url.URL {
	url, err := r.URL(pairs...)
	if err != nil {
		log.Fatalf("Error r.URL failed with error: %v", err)
	}
	return url
}

// Middleware (this function) makes adding more than one layer of middleware easy
// by specifying them as a list. It will run the last specified handler first.
func Middleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}
