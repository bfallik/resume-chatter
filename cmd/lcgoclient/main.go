package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
)

func ask() error {
	ctx := context.Background()

	llm, err := openai.New()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	cmd := exec.Command("pdftotext", "/home/bfallik/Documents/JobSearches/bfallik-resume/bfallik-resume.pdf", "-")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return err
	}

	loader := documentloaders.NewText(buf)
	docs, err := loader.LoadAndSplit(ctx, textsplitter.NewRecursiveCharacter())
	if err != nil {
		return fmt.Errorf("split: %v", err)
	}

	// TODO - find similar docs

	stuffQAChain := chains.LoadStuffQA(llm)
	answer, err := chains.Call(context.Background(), stuffQAChain, map[string]any{
		"input_documents": docs,
		"question":        "Where did Brian go to collage?",
	})
	if err != nil {
		return err
	}
	log.Println(answer)

	return nil
}

func main() {
	if err := ask(); err != nil {
		log.Fatalln(err)
	}
}
