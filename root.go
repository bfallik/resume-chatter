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

	"github.com/a-h/templ/cmd/templ/generatecmd/sse"
	"github.com/bfallik/resume-chatter/internal/model"
	"github.com/bfallik/resume-chatter/views/components"
	"github.com/bfallik/resume-chatter/views/pages"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

//go:embed static/**
var staticFS embed.FS

var chatHistory []model.Chat = []model.Chat{
	{
		IsStart:   true,
		Header:    "Obi-Wan Kenobi",
		IsWaiting: false,
		Bubble:    "You were the Chosen One!",
	},
	{
		IsStart:   false,
		Header:    "Anakin",
		IsWaiting: false,
		Bubble:    "I loved you.",
	},
}

func Serve(address string) error {
	start := time.Now()
	log.Printf("started %v", start.Format(time.RFC1123))

	ctx := context.Background()

	llm, err := openai.New()
	if err != nil {
		return err
	}

	server := sse.New()

	buf := new(bytes.Buffer)
	cmd := exec.Command("pdftotext", "/home/bfallik/Documents/JobSearches/bfallik-resume/bfallik-resume.pdf", "-")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return err
	}

	loader := documentloaders.NewText(buf)
	docs, err := loader.LoadAndSplit(ctx, textsplitter.NewRecursiveCharacter())
	if err != nil {
		return err
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Handle("/static/*", http.FileServer(http.FS(staticFS)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		idx := pages.Index(chatHistory, start)
		err := idx.Render(r.Context(), w)
		if err != nil {
			log.Printf("err rendering html template: %+v\n", err)
			http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
		}
	})

	r.Post("/ask", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("err parsing form: %+v\n", err)
			http.Error(w, "error parsing form", http.StatusInternalServerError)
			return
		}

		content, ok := r.Form["content"]
		if !ok {
			log.Printf("missing form value: content\n")
			http.Error(w, "missing form value: content", http.StatusInternalServerError)
			return
		}

		question := content[0] // BF TODO: handle this
		log.Println("question: ", question)
		chatHistory = append(chatHistory,
			model.Chat{
				IsStart:   true,
				Header:    "Obi-Wan Kenobi",
				IsWaiting: false,
				Bubble:    question,
			},
			model.Chat{
				IsStart:   false,
				Header:    "Anakin",
				IsWaiting: true,
				Bubble:    "",
			})

		// BF TODO: this is super racy!
		go func() {
			time.Sleep(2 * time.Second)

			// TODO - find similar docs

			stuffQAChain := chains.LoadStuffQA(llm)
			answer, err := chains.Call(context.Background(), stuffQAChain, map[string]any{
				"input_documents": docs,
				"question":        question,
			})
			if err != nil {
				log.Printf("LoadStuffQA: %+v\n", err)
				return
			}

			chatHistory[len(chatHistory)-1].IsWaiting = false
			chatHistory[len(chatHistory)-1].Bubble = fmt.Sprintf("%v", answer["text"])
			log.Println("answer: ", answer)

			server.Send("chat-events", answer["text"].(string))
		}()

		if err := components.Chat(chatHistory).Render(r.Context(), w); err != nil {
			log.Printf("err rendering html template: %+v\n", err)
			http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
		}
	})

	r.Get("/chat-history", func(w http.ResponseWriter, r *http.Request) {
		if err := components.Chat(chatHistory).Render(r.Context(), w); err != nil {
			log.Printf("err rendering html template: %+v\n", err)
			http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
		}
	})

	r.Get("/events", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			<-r.Context().Done()
			//	SSE client disconnected
		}()

		server.ServeHTTP(w, r)
	})

	log.Println("webserver listening on", address)
	return http.ListenAndServe(address, r)
}
