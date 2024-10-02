package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello film-sync!")
}

func authHandler(w http.ResponseWriter, r *http.Request, acr chan bool) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Missing code or state", http.StatusUnauthorized)
		return
	}

	config, err := google.Config()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	database.SaveToken(tok)
	acr <- true
	acr <- true

	fmt.Fprintln(w, "Thank you! You can now close this tab.")
}

func newRouter(acr chan bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		authHandler(w, r, acr)
	})
	mux.HandleFunc("/daily", dailyHandler)

	return mux
}

func Listen(acr chan bool) error {
	port, err := util.LoadEnvVar("PORT")
	if err != nil {
		return err
	}
	router := newRouter(acr)

	log.Default().Printf("[HTTP] Server listening on port %s", port)

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}

	return nil
}
