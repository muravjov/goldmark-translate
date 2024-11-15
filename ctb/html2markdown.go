package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"git.catbo.net/muravjov/go2023/llmrequest"
	"git.catbo.net/muravjov/go2023/util"
)

func html2markdown(llmProvider string, logRequests bool, args []string) bool {
	if len(args) != 2 {
		util.Errorf("html2markdown: strictly 2 arguments required")
		return false
	}

	srcFilename, dstFilename := args[0], args[1]

	var html string
	if srcFilename == "-" {
		dat, err := io.ReadAll(os.Stdin)
		if err != nil {
			util.Errorf("error while reading file %v: %v", srcFilename, err)
			return false
		}
		html = string(dat)
	} else {
		dat, err := os.ReadFile(srcFilename)
		if err != nil {
			util.Errorf("error while reading file %v: %v", srcFilename, err)
			return false
		}
		html = string(dat)
	}

	var dstF *os.File
	if dstFilename == "-" {
		dstF = os.Stdout
	} else {
		f, err := os.Open(dstFilename)
		if err != nil {
			util.Errorf("error while opening file %v: %v", dstFilename, err)
			return false
		}
		defer f.Close()

		dstF = f
	}

	client, err := llmrequest.MakeClient(llmProvider, logRequests)
	if err != nil {
		return false
	}

	stream, err := llmrequest.HTML2Markdown(client, html)
	if err != nil {
		util.Errorf("ChatCompletionStream error: %v\n", err)
		return false
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			util.Errorf("\nStream error: %v\n", err)
			return false
		}

		fmt.Fprintf(dstF, response.Choices[0].Delta.Content)
	}

	return true
}
