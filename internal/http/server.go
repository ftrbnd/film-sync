package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
)

var ctx context.Context
var googleConfig *oauth2.Config

func Listen(c context.Context, config *oauth2.Config, f func(l []string) error) error {
	port, err := util.LoadEnvVar("PORT")
	if err != nil {
		return err
	}

	ctx = c
	googleConfig = config

	router := newRouter(f)

	log.Default().Printf("[HTTP] Server listening on port %s", port)
	log.Default().Println("[Film Sync] All services ready") // server is the last service to start since it blocks

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}

	return nil
}

func newRouter(f func(l []string) error) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/auth", authHandler)
	mux.HandleFunc("/daily", func(w http.ResponseWriter, r *http.Request) {
		dailyHandler(w, r, f)
	})
	mux.HandleFunc("/api/scans", scansHandler)
	mux.HandleFunc("/api/scans/{folder}", scanHandler)

	return mux
}
