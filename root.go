package resume_chatter

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/bfallik/resume-chatter/internal/model"
	"github.com/bfallik/resume-chatter/views/components"
	"github.com/bfallik/resume-chatter/views/pages"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

//go:embed static/**
var staticFS embed.FS

type History struct {
	data []model.ChatMessage
	mu   sync.RWMutex
}

func (h *History) GetChat() []model.ChatMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.data
}

func (h *History) Append(cs ...model.ChatMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data = append(h.data, cs...)
}

func (h *History) UpdateWaiting(newBubble string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ln := len(h.data)
	h.data[ln-1].IsWaiting = false
	h.data[ln-1].Bubble = newBubble
}

type LLMAlert struct {
	llmErr error
	mu     sync.RWMutex
}

func (a *LLMAlert) GetErr() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.llmErr
}

func (a *LLMAlert) SetErr(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.llmErr = err
}

type Server struct {
	Start       time.Time
	LLM         *openai.LLM
	Docs        []schema.Document
	ChatHistory History
	Alert       LLMAlert
}

func NewServer(ctx context.Context) (*Server, error) {
	start := time.Now()
	log.Printf("started %v", start.Format(time.RFC1123))

	llm, err := openai.New()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	cmd := exec.Command("pdftotext", "/home/bfallik/Documents/JobSearches/bfallik-resume/bfallik-resume.pdf", "-")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	loader := documentloaders.NewText(buf)
	docs, err := loader.LoadAndSplit(ctx, textsplitter.NewRecursiveCharacter())
	if err != nil {
		return nil, err
	}

	return &Server{
		Start: start,
		LLM:   llm,
		Docs:  docs,
		ChatHistory: History{
			data: []model.ChatMessage{
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
			},
		},
	}, nil
}

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	idx := pages.Index(s.ChatHistory.GetChat(), s.Start)
	err := idx.Render(r.Context(), w)
	if err != nil {
		log.Printf("err rendering html template: %+v\n", err)
		http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
	}
}

func (s *Server) AskHandler(w http.ResponseWriter, r *http.Request) {
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
	ln := len(s.ChatHistory.GetChat())
	newMsgs := []model.ChatMessage{
		{
			IsStart:   true,
			Header:    "Obi-Wan Kenobi",
			IsWaiting: false,
			Bubble:    question,
		},
		{
			IsStart:             false,
			Header:              "Anakin",
			IsWaiting:           true,
			WaitingMessageIndex: ln + 1,
			Bubble:              "",
		},
	}
	s.ChatHistory.Append(newMsgs...)

	go func() {
		time.Sleep(2 * time.Second)

		// TODO - find similar docs

		stuffQAChain := chains.LoadStuffQA(s.LLM)
		answer, err := chains.Call(context.Background(), stuffQAChain, map[string]any{
			"input_documents": s.Docs,
			"question":        question,
		})
		if err != nil {
			log.Printf("error calling chain: %+v\n", err)
			s.Alert.SetErr(errors.New("error calling LLM"))
			return
		}

		s.ChatHistory.UpdateWaiting(fmt.Sprintf("%v", answer["text"]))
		log.Println("answer: ", answer)
	}()

	if err := components.ChatHistoryElements(newMsgs...).Render(r.Context(), w); err != nil {
		log.Printf("err rendering html template: %+v\n", err)
		http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
	}
}

func (s *Server) MessageHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := r.URL.Query().Get("index")
	n, err := strconv.Atoi(idxStr)
	if err != nil {
		log.Printf("err converting index '%v': %+v\n", idxStr, err)
		http.Error(w, "error converting index", http.StatusBadRequest)
		return
	}

	h := s.ChatHistory.GetChat()
	ln := len(h)
	if n >= ln {
		log.Printf("err converting index '%v': %+v\n", idxStr, err)
		http.Error(w, "error converting index", http.StatusNotFound)
		return
	}

	if err := components.ChatHistoryElements(h[n]).Render(r.Context(), w); err != nil {
		log.Printf("err rendering html template: %+v\n", err)
		http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
		return
	}

	alertErr := s.Alert.GetErr()
	if alertErr != nil {
		if err := components.Alert(model.Alert{MsgText: alertErr.Error()}).Render(r.Context(), w); err != nil {
			log.Printf("err rendering html template: %+v\n", err)
			http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
		}
	}
}

func (s *Server) Serve(address string) error {
	r := chi.NewRouter()

	// middleware

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// routes

	r.Handle("/static/*", http.FileServer(http.FS(staticFS)))

	r.Get("/", s.RootHandler)
	r.Post("/ask", s.AskHandler)
	r.Get("/message", s.MessageHandler)

	log.Println("webserver listening on", address)
	return http.ListenAndServe(address, r)
}
