package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"git.catbo.net/muravjov/go2023/llmrequest"
	"git.catbo.net/muravjov/go2023/util"
)

func openSrc(srcFilename string) ([]byte, bool) {
	if srcFilename == "-" {
		dat, err := io.ReadAll(os.Stdin)
		if err != nil {
			util.Errorf("error while reading file %v: %v", srcFilename, err)
			return nil, false
		}
		return dat, true
	}

	dat, err := os.ReadFile(srcFilename)
	if err != nil {
		util.Errorf("error while reading file %v: %v", srcFilename, err)
		return nil, false
	}
	return dat, true
}

func openDst(dstFilename string) (*os.File, bool) {
	if dstFilename == "-" {
		return os.Stdout, true
	}

	f, err := os.Open(dstFilename)
	if err != nil {
		util.Errorf("error while opening file %v: %v", dstFilename, err)
		return nil, false
	}
	return f, true
}

func html2markdown(llmProvider string, logRequests bool, args []string) bool {
	if len(args) != 2 {
		util.Errorf("html2markdown: strictly 2 arguments required")
		return false
	}

	srcFilename, dstFilename := args[0], args[1]

	dat, res := openSrc(srcFilename)
	if !res {
		return res
	}
	html := string(dat)

	dstF, res := openDst(dstFilename)
	if !res {
		return res
	}
	defer dstF.Close()

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
