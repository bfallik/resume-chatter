package main

import (
	"context"
	"log/slog"

	resume_chatter "github.com/bfallik/resume-chatter"
)

func main() {
	ctx := context.Background()
	const a = ":8080"

	svr, err := resume_chatter.NewServer(ctx)
	if err != nil {
		slog.Error("creating server", slog.Any("error", err))
	}

	if err := svr.Serve(a); err != nil {
		slog.Error("serve", slog.Any("error", err))
	}
}
