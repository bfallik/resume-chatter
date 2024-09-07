package main

import (
	"log"

	rc "github.com/bfallik/resume-chatter"
)

func main() {
	const a = ":8080"
	log.Fatalln(rc.Serve(a))
}
