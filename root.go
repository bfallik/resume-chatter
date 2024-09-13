package root

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/bfallik/resume-chatter/internal/model"
	chatv1 "github.com/bfallik/resume-chatter/protocgengo/chat/v1"
	"github.com/bfallik/resume-chatter/views/components"
	"github.com/bfallik/resume-chatter/views/pages"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

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

var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello", // BF TODO
}

type GRPCClient struct{ client chatv1.ChatServiceClient }

func (m *GRPCClient) Ask(ctx context.Context, in *chatv1.AskRequest, opts ...grpc.CallOption) (*chatv1.AskResponse, error) {
	return m.client.Ask(ctx, in, opts...)
}

// LLM is the interface that we're exposing as a plugin.
type LLM interface {
	Ask(request chatv1.AskRequest) (chatv1.AskResponse, error)
}

type LLMGRPCPlugin struct {
	plugin.Plugin
	Impl LLM
}

func (p *LLMGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	panic("NOT IMPLEMENTED")
}

func (p *LLMGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: chatv1.NewChatServiceClient(c)}, nil
}

const LLMKey = "llm_grpc"

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	LLMKey: &LLMGRPCPlugin{},
}

func Serve(address string) error {
	start := time.Now()
	log.Printf("started %v", start.Format(time.RFC1123))

	llmPlugin := os.Getenv("LLM_PLUGIN")
	if len(llmPlugin) == 0 {
		log.Fatalln("missing LLM_PLUGIN env var")
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          PluginMap,
		Cmd:              exec.Command("sh", "-c", os.Getenv("LLM_PLUGIN")),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger: hclog.FromStandardLogger(
			log.Default(),
			&hclog.LoggerOptions{Name: "plugin", Level: hclog.Debug},
		),
	})
	defer client.Kill()

	addr, err := client.Start()
	if err != nil {
		log.Fatalf("plugin start error: %v", err)
	}
	log.Println("plugin listening on", addr)

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

		// Connect via RPC
		rpcClient, err := client.Client()
		if err != nil {
			log.Printf("unable to get client: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to get client"))
			return
		}

		// Request the plugin
		raw, err := rpcClient.Dispense(LLMKey)
		if err != nil {
			log.Printf("unable to dispense: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to dispense"))
			return
		}

		chatService := raw.(chatv1.ChatServiceClient)
		response, err := chatService.Ask(r.Context(), &chatv1.AskRequest{
			DocumentPath: "/home/bfallik/Documents/JobSearches/bfallik-resume/bfallik-resume.pdf",
			Question:     question,
		})
		if err != nil {
			log.Printf("chat service error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("chat service error"))
			return
		}

		chatHistory = append(chatHistory, model.Chat{
			IsStart: false,
			Header:  "Anakin",
			Bubble:  response.Response,
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
