package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"net/http"
	"os"

	"caos-test-op/mock"

	"github.com/caos/oidc/pkg/op"
	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()

	var port, issuer string

	if port = os.Getenv("DUMMY_OIDC_PORT"); port == "" {
		port = "9998"
	}
	if issuer = os.Getenv("DUMMY_OIDC_ISSUER"); issuer == "" {
		issuer = "http://dummy-oidc:9998"
	}

	b := make([]byte, 32)
	rand.Read(b)
	cryptoKey := sha256.Sum256(b)

	config := &op.Config{
		Issuer:         issuer,
		CryptoKey:      cryptoKey,
		CodeMethodS256: true,
	}
	storage := mock.NewAuthStorage()

	handler, err := op.NewOpenIDProvider(ctx, config, storage)

	if err != nil {
		log.Fatal(err)
	}

	router := handler.HttpHandler().(*mux.Router)
	router.Methods("GET").Path("/login").HandlerFunc(HandleLogin)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	err = server.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	requestId := r.Form.Get("id")

	// simulate user login and retrieve a token that indicates a successfully logged-in user
	usertoken := requestId + ":usertoken"

	http.Redirect(w, r, "/authorize/callback?id="+usertoken, http.StatusFound)
}
