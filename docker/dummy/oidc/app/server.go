// Command dummy-oidc starts a minimal OIDC provider used by docker-compose for local development.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zitadel/oidc/v3/pkg/op"

	"caos-test-op/mock"
)

const (
	defaultIssuer = "http://dummy-oidc:9998"
	defaultPort   = "9998"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	handler, err := newHandler(defaultIssuer, mock.NewAuthStorage())
	if err != nil {
		log.Printf("failed to create OIDC provider: %v", err)
		return
	}

	server := &http.Server{
		Addr:              ":" + defaultPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("OIDC server stopped with error: %v", err)
	}
}

// newHandler returns the HTTP handler that serves the provider endpoints and the dummy login.
func newHandler(issuer string, storage *mock.AuthStorage) (http.Handler, error) {
	provider, err := newProvider(issuer, storage)
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()
	loginHandler := newLoginHandler(storage, op.AuthCallbackURL(provider), op.NewIssuerInterceptor(provider.IssuerFromRequest))
	router.Get("/login", loginHandler)
	router.Mount("/", provider)

	return router, nil
}

// newProvider builds an OpenID provider with the dummy's permissive defaults.
func newProvider(issuer string, storage op.Storage) (*op.Provider, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	cfg := &op.Config{
		CryptoKey:               sha256.Sum256(b),
		CodeMethodS256:          true,
		AuthMethodPost:          true,
		AuthMethodPrivateKeyJWT: true,
		GrantTypeRefreshToken:   true,
		RequestObjectSupported:  true,
	}
	return op.NewProvider(cfg, storage, requestIssuer(issuer),
		op.WithAllowInsecure(),
	)
}

// requestIssuer returns an issuer function that answers "https://<host>" for requests a reverse
// proxy received over HTTPS, so clients that require an https issuer can use the dummy through
// Traefik, and the static default issuer for all other requests.
func requestIssuer(defaultIssuer string) func(bool) (op.IssuerFromRequest, error) {
	return func(bool) (op.IssuerFromRequest, error) {
		return func(r *http.Request) string {
			if !strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
				return defaultIssuer
			}

			host := r.Header.Get("X-Forwarded-Host")

			if host == "" {
				host = r.Host
			}

			if host == "" || strings.ContainsAny(host, "/\\?#@ ") {
				return defaultIssuer
			}

			return "https://" + host
		}, nil
	}
}

// newLoginHandler returns the dummy /login handler. It marks the auth request as
// authenticated and redirects back to the OP's auth callback so the code flow
// completes without any user interaction.
func newLoginHandler(storage *mock.AuthStorage, callback func(context.Context, string) string, issuerInterceptor *op.IssuerInterceptor) http.HandlerFunc {
	return issuerInterceptor.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleLogin(w, r, storage, callback)
	})
}

// HandleLogin marks an in-flight auth request as authenticated and redirects to the OP callback.
// It is exported to keep the existing test surface stable.
func HandleLogin(w http.ResponseWriter, r *http.Request, storage *mock.AuthStorage, callback func(context.Context, string) string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid login request", http.StatusBadRequest)
		return
	}
	id := r.Form.Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	if err := storage.MarkRequestDone(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, callback(r.Context(), id), http.StatusFound)
}
