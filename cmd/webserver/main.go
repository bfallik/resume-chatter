package main

import (
	"context"
	"log"

	rc "github.com/bfallik/resume-chatter"
)

func main() {
	ctx := context.Background()
	const a = ":8080"

	svr, err := rc.NewServer(ctx)
	if err != nil {
		log.Fatalln("error creating server: ", err)
	}

	log.Fatalln(svr.Serve(a))
}
