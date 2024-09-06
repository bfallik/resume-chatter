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
	defaultQuestion = "What is the airspeed velocity of an unladen sparrow?"
)

var (
	addr     = flag.String("addr", "localhost:8081", "the address to connect to")
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Ask(ctx, &chatv1.AskRequest{Question: *question})
	if err != nil {
		log.Fatalf("could not ask: %v", err)
	}
	log.Printf("Response: %s", r.GetResponse())
}
