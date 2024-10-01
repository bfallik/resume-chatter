package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	resume_chatter "github.com/bfallik/resume-chatter"
	"github.com/cue-exp/cueconfig"
)

type webserverConfig struct {
	Port int `json:"port"`
}

type webserverRuntime struct {
}

//go:embed schema.cue
var schema []byte

//go:embed defaults.cue
var defaults []byte

func main() {
	ctx := context.Background()

	configFile := "config/webserver.cue"

	// This is a placeholder for any runtime values provided
	// as input to the configuration.
	runtime := struct {
		Runtime webserverRuntime `json:"runtime"`
	}{
		Runtime: webserverRuntime{},
	}
	var cfg webserverConfig
	if err := cueconfig.Load(configFile, schema, defaults, runtime, &cfg); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	svr, err := resume_chatter.NewServer(ctx)
	if err != nil {
		slog.Error("creating server", slog.Any("error", err))
	}

	if err := svr.Serve(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		slog.Error("serve", slog.Any("error", err))
	}
}
