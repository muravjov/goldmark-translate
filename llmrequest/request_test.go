package llmrequest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"git.catbo.net/muravjov/go2023/util"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
	"moul.io/http2curl"
)

func getGGCAccessToken() (string, error) {

	url := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(util.Map2URLPath(map[string]string{
		"scope": "GIGACHAT_API_PERS",
	})))

	if err != nil {
		util.Errorf("http.NewRequest failed: %v", err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("RqUID", uuid.New().String())
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v",
		base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%v:%v", os.Getenv("GIGACHAT_CLIENT_ID"), os.Getenv("GIGACHAT_CLIENT_SECRET"))),
		),
	),
	)

	res, err := client.Do(req)
	if err != nil {
		util.Errorf("fetching access token failed: %v", err)
		return "", err
	}

	if err := util.CheckStatusCodeIs2XX(res); err != nil {
		return "", err
	}

	defer res.Body.Close()

	v := &struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		util.Errorf("access token decoding error: %v", err)
		return "", err
	}

	return v.AccessToken, nil
}

type RoundTripperFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRequest(t *testing.T) {
	//t.SkipNow()

	const (
		llmOpenai   = "openai"
		llmGigachat = "gigachat"
	)

	llmProvider := llmGigachat // llmOpenai //
	debugFlag := true          // false //

	wrapTransport := func(tr http.RoundTripper) http.RoundTripper {
		if !debugFlag {
			return tr
		}

		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			command, _ := http2curl.GetCurlCommand(req)
			fmt.Println(command)

			return tr.RoundTrip(req)
		})
	}

	var client *openai.Client
	var llmModel string

	switch llmProvider {
	case llmOpenai:
		config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))

		if proxyURL := os.Getenv("OPENAI_HTTP_PROXY"); proxyURL != "" {
			u, err := url.Parse(proxyURL)
			if err != nil {
				panic(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(u),
			}
			config.HTTPClient = &http.Client{
				Transport: wrapTransport(transport),
			}
		}

		client = openai.NewClientWithConfig(config)
		llmModel = openai.GPT3Dot5Turbo
	case llmGigachat:
		ggcToken, err := getGGCAccessToken()
		if err != nil {
			panic(err)
		}

		config := openai.DefaultConfig(ggcToken)

		// https://developers.sber.ru/docs/ru/gigachat/api/reference/rest/post-chat
		config.BaseURL = "https://gigachat.devices.sberbank.ru/api/v1"
		config.HTTPClient = &http.Client{
			Transport: wrapTransport(&http.Transport{}),
		}

		client = openai.NewClientWithConfig(config)

		const (
			GigaChatLite = "GigaChat"
			GigaChatPro  = "GigaChat-Pro"
		)

		llmModel = GigaChatPro
	default:
		panic(fmt.Sprintf("unknown llm provider: %v", llmProvider))
	}

	if false {
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: llmModel,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "Hello!",
					},
				},
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}

		fmt.Println(resp.Choices[0].Message.Content)
	}

	if true {
		req := openai.ChatCompletionRequest{
			Model:     llmModel,
			MaxTokens: 20,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Lorem ipsum",
				},
			},
			Stream: true,
		}
		stream, err := client.CreateChatCompletionStream(context.Background(), req)
		if err != nil {
			fmt.Printf("ChatCompletionStream error: %v\n", err)
			return
		}
		defer stream.Close()

		fmt.Printf("Stream response: ")
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println("\nStream finished")
				return
			}

			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				return
			}

			fmt.Printf(response.Choices[0].Delta.Content)
		}

	}

}
