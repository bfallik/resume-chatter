package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bfallik/resume-chatter/internal/model"
	"github.com/bfallik/resume-chatter/views/pages"
)

var chatHistory []model.ChatMessage = []model.ChatMessage{
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

func main() {
	start := time.Now()

	if err := pages.Index(chatHistory, start, model.Alert{}).Render(context.Background(), os.Stdout); err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}
}
