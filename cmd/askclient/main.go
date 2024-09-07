package main

import (
	"context"
	"flag"
	"log"
	"time"

	chatv1 "github.com/bfallik/resume-chatter/protocgengo/chat/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultQuestion = "What was Brian's second most recent job and when did he work there?"
)

var (
	addr     = flag.String("addr", "localhost:8081", "the address to connect to")
	doc_path = flag.String("document_path", "/home/bfallik/Documents/JobSearches/bfallik-resume/bfallik-resume.pdf", "Reference document to use")
	question = flag.String("question", defaultQuestion, "Question to ask")
)

func main() {
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := chatv1.NewChatServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := c.Ask(ctx, &chatv1.AskRequest{Question: *question, DocumentPath: *doc_path})
	if err != nil {
		log.Fatalf("could not ask: %v", err)
	}
	log.Printf("Response: %s", r.GetResponse())
}
