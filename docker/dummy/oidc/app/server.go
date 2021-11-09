package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"caos-test-op/mock"
	"github.com/caos/oidc/pkg/op"
)

func main() {
	ctx := context.Background()

	b := make([]byte, 32)
	rand.Read(b)

	port := "9998"
	config := &op.Config{
		Issuer:         "http://dummy-oidc:9998",
		CryptoKey:      sha256.Sum256(b),
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
