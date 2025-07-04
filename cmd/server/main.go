package main

import (
	"github.com/hugmouse/scan24/internal/handler"
	"github.com/hugmouse/scan24/static"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.IndexHandler)
	mux.HandleFunc("/analyze", handler.AnalyzeHandler)
	mux.HandleFunc("/result", handler.ResultHandler)
	mux.HandleFunc("/status", handler.JobStatus)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))

	log.Println("Starting server on :8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
