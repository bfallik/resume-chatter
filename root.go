package root

import (
	"embed"
	"log"
	"net/http"
	"time"

	"github.com/bfallik/resume-chatter/internal/model"
	"github.com/bfallik/resume-chatter/views/components"
	"github.com/bfallik/resume-chatter/views/pages"

	"github.com/go-chi/chi/v5"
)

//go:embed static/**
var staticFS embed.FS

var chatHistory []model.Chat = []model.Chat{
	{
		IsStart: true,
		Header:  "Obi-Wan Kenobi",
		Bubble:  "You were the Chosen One!",
	},
	{
		IsStart: false,
		Header:  "Anakin",
		Bubble:  "I loved you.",
	},
}

func Serve(address string) error {
	start := time.Now()
	log.Printf("started %v", start.Format(time.RFC1123))

	r := chi.NewRouter()

	r.Handle("/static/*", http.FileServer(http.FS(staticFS)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		idx := pages.Index(chatHistory, start)
		err := idx.Render(r.Context(), w)
		if err != nil {
			log.Printf("err rendering html template: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error rendering HTML template"))
		}
	})

	r.Post("/ask", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("err parsing form: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error parsing form"))
			return
		}

		content, ok := r.Form["content"]
		if !ok {
			log.Printf("missing form value: content\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("missing form value: content"))
			return
		}

		chatHistory = append(chatHistory, model.Chat{
			IsStart: true,
			Header:  "Obi-Wan Kenobi",
			Bubble:  content[0], // BF TODO: handle this
		})

		idx := components.Chat(chatHistory)
		err := idx.Render(r.Context(), w)
		if err != nil {
			log.Printf("err rendering html template: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error rendering HTML template"))
		}
	})

	log.Println("listening on ", address)
	return http.ListenAndServe(address, r)
}
