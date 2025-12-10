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
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zitadel/oidc/pkg/op"

	"caos-test-op/mock"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	b := make([]byte, 32)
	_, _ = rand.Read(b)
	if _, err := rand.Read(b); err != nil {
		log.Printf("failed to seed crypto key: %v", err)
		return
	}

	port := "9998"
	config := &op.Config{
		Issuer:         "http://dummy-oidc:9998",
		CryptoKey:      sha256.Sum256(b),
		CodeMethodS256: true,
	}
	storage := mock.NewAuthStorage()

	handler, err := op.NewOpenIDProvider(ctx, config, storage)
	if err != nil {
		log.Printf("failed to create OIDC provider: %v", err)
		return
	}
	router := handler.HttpHandler().(*mux.Router)
	router.Methods("GET").Path("/login").HandlerFunc(HandleLogin)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
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

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// HandleLogin mocks a login page and immediately redirects with a user token.
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid login request", http.StatusBadRequest)
		return
	}

	requestId := r.Form.Get("id")
	// simulate user login and retrieve a token that indicates a successfully logged-in user
	userToken := requestId + ":usertoken"

	http.Redirect(w, r, "/authorize/callback?id="+userToken, http.StatusFound)
}
