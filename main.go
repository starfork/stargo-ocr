package main

import (
	"net/http"
	"os"

	"github.com/starfork/stargo-ocrserver/controllers"
)

func main() {
	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		controllers.FileUpload(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
