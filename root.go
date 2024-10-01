package resume_chatter

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
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

func (a *LLMAlert) Model() model.Alert {
	if a.GetErr() == nil {
		return model.Alert{}
	}
	return model.Alert{
		MsgText: a.llmErr.Error(),
	}
}

func (a *LLMAlert) SetErr(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.llmErr = err
}

type LLM struct {
	LLM  *openai.LLM
	Docs []schema.Document
}

func NewLLM(ctx context.Context) (*LLM, error) {
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

	return &LLM{
		LLM:  llm,
		Docs: docs,
	}, nil
}

func (l *LLM) Call(ctx context.Context, question string) (map[string]any, error) {
	// TODO - find similar docs

	stuffQAChain := chains.LoadStuffQA(l.LLM)
	return chains.Call(ctx, stuffQAChain, map[string]any{
		"input_documents": l.Docs,
		"question":        question,
	})
}

type Server struct {
	Start       time.Time
	LLM         *LLM
	Docs        []schema.Document
	ChatHistory History
	Alert       LLMAlert
}

func NewServer(ctx context.Context) (*Server, error) {
	start := time.Now()
	slog.Info("server", "start", start.Format(time.RFC1123))

	llm, err := NewLLM(ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		Start: start,
		LLM:   llm,
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
	idx := pages.Index(s.ChatHistory.GetChat(), s.Start, s.Alert.Model())
	if err := idx.Render(r.Context(), w); err != nil {
		slog.Error("rendering html template: ", slog.Any("error", err))
		http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
	}
}

func (s *Server) AskHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("parsing form: ", slog.Any("error", err))
		http.Error(w, "error parsing form", http.StatusInternalServerError)
		return
	}

	content, ok := r.Form["content"]
	if !ok {
		slog.Error("missing form value", slog.Any("content", content))
		http.Error(w, "missing form value: content", http.StatusInternalServerError)
		return
	}

	question := content[0] // BF TODO: handle this
	slog.Info("", "question", question)
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

	if err := components.ChatHistoryElements(newMsgs...).Render(r.Context(), w); err != nil {
		slog.Error("chat history render", slog.Any("error", err))
		http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
	}
}

func (s *Server) MessageHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := r.URL.Query().Get("index")
	n, err := strconv.Atoi(idxStr)
	if err != nil {
		slog.Error("index out of range", slog.String("index", idxStr), slog.Any("error", err))
		http.Error(w, "index out of range", http.StatusBadRequest)
		return
	}

	h := s.ChatHistory.GetChat()
	ln := len(h)
	if n >= ln {
		slog.Error("out of bounds", slog.Int("index", n), slog.Int("len", ln))
		http.Error(w, "error converting index", http.StatusNotFound)
		return
	}

	answer, err := s.LLM.Call(context.Background(), h[n-1].Bubble)
	if err != nil {
		slog.Error("LLM chain call", slog.Any("error", err))
		http.Error(w, "error calling LLM", http.StatusInternalServerError)
		s.Alert.SetErr(errors.New("error calling LLM"))
		return
	}

	s.ChatHistory.UpdateWaiting(fmt.Sprintf("%v", answer["text"]))
	slog.Info("", "answer", answer)

	if err := components.ChatHistoryElements(h[n]).Render(r.Context(), w); err != nil {
		slog.Error("chat history render", slog.Any("error", err))
		http.Error(w, "error rendering HTML template", http.StatusInternalServerError)
		return
	}

	alertErr := s.Alert.GetErr()
	if alertErr != nil {
		if err := components.Alert(s.Alert.Model()).Render(r.Context(), w); err != nil {
			slog.Error("alert render", slog.Any("error", err))
			http.Error(w, "error rendering alert template", http.StatusInternalServerError)
		}
	}
}

func (s *Server) DismissHandler(w http.ResponseWriter, r *http.Request) {
	s.Alert.SetErr(nil)
	w.WriteHeader(http.StatusOK)
	if err := components.Alert(s.Alert.Model()).Render(r.Context(), w); err != nil {
		slog.Error("alert render", slog.Any("error", err))
		http.Error(w, "error rendering alert template", http.StatusInternalServerError)
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
	r.Post("/dismiss", s.DismissHandler)

	slog.Info("webserver listening on", "address", address)
	return http.ListenAndServe(address, r)
}
