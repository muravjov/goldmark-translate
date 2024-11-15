package llmrequest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"git.catbo.net/muravjov/go2023/util"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
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

func unknownLLMProvider(llmProvider string) {
	panic(fmt.Sprintf("unknown llm provider: %v", llmProvider))
}

const (
	llmOpenai   = "openai"
	llmGigachat = "gigachat"
)

type Client struct {
	Client      *openai.Client
	LLMProvider string
}

func MakeClient(llmProvider string, logRequests bool) (*Client, error) {
	wrapTransport := func(tr http.RoundTripper) http.RoundTripper {
		if !logRequests {
			return tr
		}

		if tr == nil {
			tr = http.DefaultTransport
		}

		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			command, _ := http2curl.GetCurlCommand(req)
			fmt.Fprintln(os.Stderr, command)

			return tr.RoundTrip(req)
		})
	}

	var client *openai.Client

	switch llmProvider {
	case llmOpenai:
		config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))

		if proxyURL := os.Getenv("OPENAI_HTTP_PROXY"); proxyURL != "" {
			u, err := url.Parse(proxyURL)
			if err != nil {
				util.Error(err)
				return nil, err
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(u),
			}
			config.HTTPClient = &http.Client{
				Transport: wrapTransport(transport),
			}
		}

		client = openai.NewClientWithConfig(config)
	case llmGigachat:
		ggcToken, err := getGGCAccessToken()
		if err != nil {
			return nil, err
		}

		config := openai.DefaultConfig(ggcToken)

		// https://developers.sber.ru/docs/ru/gigachat/api/reference/rest/post-chat
		config.BaseURL = "https://gigachat.devices.sberbank.ru/api/v1"
		config.HTTPClient = &http.Client{
			Transport: wrapTransport(nil),
		}

		client = openai.NewClientWithConfig(config)
	default:
		unknownLLMProvider(llmProvider)
	}

	return &Client{
		Client:      client,
		LLMProvider: llmProvider,
	}, nil
}

const (
	GigaChatLite       = "GigaChat"
	GigaChatPro        = "GigaChat-Pro"
	GigaChatEmbeddings = "Embeddings"
)

func (c *Client) GetLLModel(fastModel bool) string {
	switch c.LLMProvider {
	case llmOpenai:
		if fastModel {
			return openai.GPT3Dot5Turbo
		}
		return openai.GPT4o
	case llmGigachat:
		if fastModel {
			return GigaChatLite
		}
		return GigaChatPro
	default:
		unknownLLMProvider(c.LLMProvider)
	}

	return ""
}

func HTML2Markdown(client *Client, html string) (stream *openai.ChatCompletionStream, err error) {
	req := openai.ChatCompletionRequest{
		Model: client.GetLLModel(false),
		//MaxTokens: 40,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Your task is to convert the following html text into markdown format:",
				//Content: "Your task is to convert the following html text into markdown format; you leave href attr and image src attr not changed. The html:",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: html,
			},
		},
		Stream:      true,
		Temperature: 0, // 0.00001, //
		TopP:        0, // 0.00001, //
	}
	return client.Client.CreateChatCompletionStream(context.Background(), req)
}
