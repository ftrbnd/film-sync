package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello film-sync!")
}

func authHandler(w http.ResponseWriter, r *http.Request, acr chan *oauth2.Token) {
	code := r.URL.Query().Get("code")
	config := google.Config()

	tok, err := config.Exchange(context.TODO(), code)
	util.CheckError("Unable to retrieve token from web", err)

	google.SaveToken("token.json", tok)
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
	port := util.LoadEnvVar("PORT")
	router := newRouter(acr)

	log.Default().Printf("Server listening on port %s", port)

	err := http.ListenAndServe(":"+port, router)
	util.CheckError("Failed to start http server", err)
}
