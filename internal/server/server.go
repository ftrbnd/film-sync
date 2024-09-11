package server

import (
	"fmt"
	"log"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello film-sync!")
}

func newRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)

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
