package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
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

func startJob(links []string) error {
	dst := "output"
	format := "tif"

	for _, link := range links {
		z, err := files.DownloadFrom(link)
		if err != nil {
			return fmt.Errorf("failed to download from link: %v", err)
		}

		files.Unzip(z, dst, format)
		c, err := files.ConvertToPNG(format, dst)
		if err != nil {
			return fmt.Errorf("failed to convert to png: %v", err)
		}

		s3Folder, driveFolderID, message, err := files.Upload(dst, z, c)
		if err != nil {
			return fmt.Errorf("failed to upload files: %v", err)
		}

		err = discord.SendSuccessMessage(s3Folder, driveFolderID, message)
		if err != nil {
			return fmt.Errorf("failed to send discord success message: %v", err)
		}
	}

	log.Default().Println("Finished running daily job!")
	return nil
}

func dailyHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("Received /daily request")
	// TODO: authorize request
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Request accepted for processing")

	go func() {
		newLinks, err := google.CheckEmail()
		if err != nil {
			discord.SendErrorMessage(err)
			return
		}

		log.Default().Printf("Found %d new links", len(newLinks))

		if len(newLinks) > 0 {
			err = startJob(newLinks)
			if err != nil {
				discord.SendErrorMessage(err)
			}
		}
	}()
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
