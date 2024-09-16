package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/gmail"
	"golang.org/x/oauth2"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello film-sync!")
}

func authHandler(w http.ResponseWriter, r *http.Request, acr chan *oauth2.Token) {
	code := r.URL.Query().Get("code")
	config := gmail.Config()

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	gmail.SaveToken("token.json", tok)
	acr <- tok

	fmt.Fprintln(w, "Thank you! You can now close this tab.")
}

func newRouter(acr chan *oauth2.Token) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		authHandler(w, r, acr)
	})

	return mux
}

func Listen(acr chan *oauth2.Token) {
	port := 3001
	addr := fmt.Sprintf(":%d", port)
	router := newRouter(acr)

	log.Default().Printf("Server listening on http://localhost%s", addr)

	err := http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatal(err)
	}

}
