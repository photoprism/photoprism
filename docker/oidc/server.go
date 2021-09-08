package main

import (
	"context"
	"crypto/sha256"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"caos-test-op/mock"
	"github.com/caos/oidc/pkg/op"
)

func main() {
	ctx := context.Background()
	port := "9998"
	config := &op.Config{
		Issuer:    "http://host.docker.internal:9998",
		CryptoKey: sha256.Sum256([]byte("test0123test0123test0123test0123")),
	}
	storage := mock.NewAuthStorage()

	//opts := []op.Option{
	//}
	//

	handler, err := op.NewOpenIDProvider(ctx, config, storage)
	if err != nil {
		log.Fatal(err)
	}
	router := handler.HttpHandler().(*mux.Router)
	router.Methods("GET").Path("/login").HandlerFunc(HandleLogin)
	//router.Methods("POST").Path("/login").HandlerFunc(HandleCallback)
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
	//tpl := `
	//<!DOCTYPE html>
	//<html>
	//	<head>
	//		<meta charset="UTF-8">
	//		<title>Login</title>
	//	</head>
	//	<body>
	//		<form method="POST" action="/login">
	//			<input name="client"/>
	//			<button type="submit">Login</button>
	//		</form>
	//	</body>
	//</html>`
	//t, err := template.New("login").Parse(tpl)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//err = t.Execute(w, nil)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//}
	http.Redirect(w, r, "/authorize/callback?id=loginId", http.StatusFound)
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	client := r.FormValue("client")
	http.Redirect(w, r, "/authorize/callback?id="+client, http.StatusFound)
}
