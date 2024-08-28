package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/bfallik/resume-chatter/views/components"

	"github.com/go-chi/chi/v5"
)

//go:embed static/**
var staticFS embed.FS

func main() {
	r := chi.NewRouter()

	r.Handle("/static/*", http.FileServer(http.FS(staticFS)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		accordion := components.AccordionExample()
		err := accordion.Render(r.Context(), w)
		if err != nil {
			log.Printf("err rendeting html template: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error rendering HTML template"))
		}
	})

	http.ListenAndServe(":8080", r)
}
