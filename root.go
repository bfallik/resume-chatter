package root

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/bfallik/resume-chatter/internal/model"
	"github.com/bfallik/resume-chatter/views/components"
	"github.com/bfallik/resume-chatter/views/pages"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"

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

		question := content[0] // BF TODO: handle this
		chatHistory = append(chatHistory, model.Chat{
			IsStart: true,
			Header:  "Obi-Wan Kenobi",
			Bubble:  question,
		})

		llm, err := openai.New()
		if err != nil {
			log.Printf("openai LLM: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("openai LLM"))
			return
		}

		buf := new(bytes.Buffer)
		cmd := exec.Command("pdftotext", "/home/bfallik/Documents/JobSearches/bfallik-resume/bfallik-resume.pdf", "-")
		cmd.Stdout = buf
		if err := cmd.Run(); err != nil {
			log.Printf("pdftotext: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("pdftotext"))
			return
		}

		loader := documentloaders.NewText(buf)
		docs, err := loader.LoadAndSplit(r.Context(), textsplitter.NewRecursiveCharacter())
		if err != nil {
			log.Printf("LoadAndSplit: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("LoadAndSplit"))
			return
		}

		// TODO - find similar docs

		stuffQAChain := chains.LoadStuffQA(llm)
		answer, err := chains.Call(context.Background(), stuffQAChain, map[string]any{
			"input_documents": docs,
			"question":        "Where did Brian go to collage?",
		})
		if err != nil {
			log.Printf("LoadStuffQA: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("LoadStuffQA"))
		}

		chatHistory = append(chatHistory, model.Chat{
			IsStart: false,
			Header:  "Anakin",
			Bubble:  fmt.Sprintf("%v", answer),
		})

		if err := components.Chat(chatHistory).Render(r.Context(), w); err != nil {
			log.Printf("err rendering html template: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error rendering HTML template"))
		}
	})

	log.Println("webserver listening on", address)
	return http.ListenAndServe(address, r)
}
