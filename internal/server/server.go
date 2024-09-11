package server

import (
	"fmt"
	"log"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello film-sync!")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	// EXAMPLE REQUEST URL:
	// http://localhost:3001/auth?state=state-token&code=4/0AQlEd8wMXWEEeO2j2gg8QeCNbXziFDXFVCxPzcs5X3APmvv1c1jdRJ9t45yvyPRJC4zEqQ&scope=https://www.googleapis.com/auth/gmail.readonly

	fmt.Fprintln(w, "AUTH")
}

func newRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/auth", authHandler)

	return mux
}

func Listen() {
	port := 3001
	addr := fmt.Sprintf(":%d", port)
	router := newRouter()

	log.Default().Printf("Server listening on http://localhost%s", addr)

	err := http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatal(err)
	}

}
