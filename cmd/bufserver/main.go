package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	connect "connectrpc.com/connect"
	chatv1 "github.com/bfallik/resume-chatter/protocgengo/chat/v1"
	"github.com/bfallik/resume-chatter/protocgengo/chat/v1/chatv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const address = "localhost:8081"

func main() {
	mux := http.NewServeMux()
	path, handler := chatv1connect.NewChatServiceHandler(&chatServiceServer{})
	mux.Handle(path, handler)
	fmt.Println("listening on", address)
	log.Fatal(http.ListenAndServe(
		address,
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	))
}

// petStoreServiceServer implements the PetStoreService API.
type chatServiceServer struct {
	chatv1connect.UnimplementedChatServiceHandler
}

// PutPet adds the pet associated with the given request into the PetStore.
func (s *chatServiceServer) Ask(
	ctx context.Context,
	req *connect.Request[chatv1.AskRequest],
) (*connect.Response[chatv1.AskResponse], error) {
	question := req.Msg.GetQuestion()
	log.Printf("Got a request to answer %s", question)
	return connect.NewResponse(&chatv1.AskResponse{}), nil
}
