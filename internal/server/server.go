package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
)

var ctx context.Context

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello film-sync!")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
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

	tok, err := config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	database.SaveToken(tok)
	google.StartServices(ctx)

	fmt.Fprintln(w, "Thank you! You can now close this tab.")
}

func scansHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("Received /scans api request")

	scans, err := database.GetScans(true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	json.NewEncoder(w).Encode(scans)
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	folder := r.PathValue("folder")
	log.Default().Printf("Received /scans/{%s} api request", folder)

	scan, err := database.GetOneScan(folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	json.NewEncoder(w).Encode(&scan)
}

func newRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/auth", authHandler)
	mux.HandleFunc("/daily", dailyHandler)
	mux.HandleFunc("/api/scans", scansHandler)
	mux.HandleFunc("/api/scans/{folder}", scanHandler)

	return mux
}

func Listen(c context.Context) error {
	port, err := util.LoadEnvVar("PORT")
	if err != nil {
		return err
	}

	ctx = c
	router := newRouter()

	log.Default().Printf("[HTTP] Server listening on port %s", port)
	log.Default().Println("[Film Sync] All services ready") // server is the last service to start since it blocks

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}

	return nil
}
