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

func unknownLLMProvider(llmProvider string) {
	panic(fmt.Sprintf("unknown llm provider: %v", llmProvider))
}

func TestRequest(t *testing.T) {
	//t.SkipNow()

	const (
		llmOpenai   = "openai"
		llmGigachat = "gigachat"
	)

	const (
		GigaChatLite       = "GigaChat"
		GigaChatPro        = "GigaChat-Pro"
		GigaChatEmbeddings = "Embeddings"
	)

	llmProvider := llmOpenai // llmGigachat //
	logRequests := true      // false //

	getLLModel := func(fastModel bool) string {
		switch llmProvider {
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
			unknownLLMProvider(llmProvider)
		}

		return ""
	}

	wrapTransport := func(tr http.RoundTripper) http.RoundTripper {
		if !logRequests {
			return tr
		}

		if tr == nil {
			tr = http.DefaultTransport
		}

		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			command, _ := http2curl.GetCurlCommand(req)
			fmt.Println(command)

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
	case llmGigachat:
		ggcToken, err := getGGCAccessToken()
		if err != nil {
			panic(err)
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

	if false {
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: getLLModel(true),
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

	if false {
		req := openai.ChatCompletionRequest{
			Model:     getLLModel(true),
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

	if true {
		req := openai.ChatCompletionRequest{
			Model: getLLModel(false),
			//MaxTokens: 40,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Your task is to convert the following html text into markdown format:",
					//Content: "Your task is to convert the following html text into markdown format; you leave href attr and image src attr not changed. The html:",
				},
				{
					Role: openai.ChatMessageRoleUser,
					Content: `<body>
<h1 class="no-num no-toc">Catbo Documentation</h1>
<h2 id="introduction"><span class="secno">1 </span>Introduction</h2>
<p> Catbo is computer-assisted translation web service, <a href="http://en.wikipedia.org/wiki/Computer-assisted_translation">CAT</a>;
it is designed to help a human translator to translate documentation and other texts. Automatic machine translation systems available today
are not able to produce high-quality translations. 
<!-- :TODO: хорошо бы завернуть про мусорный перевод автопереводчиков, но не нашел прочного аналога в английском -->
</p><p> Catbo has full support for texts in the formats: <b>HTML PDF EPUB SRT</b>. 
Our next goal is to support wikipedia/mediawiki format, odt/docx and so on (xml/docbook/rst/...).

</p><h2 id="projects"><span class="secno">2 </span>Projects</h2>
<p> Projects are used to setup a <dfn id="language-pair">language pair</dfn> for all documents to translate in them. For example, the pair English =&gt; Russian is to translate
an English text to Russian. <span>Translation Memory</span> and <span>Term Base</span> are local to a project, too.

</p><h3 id="create-a-project"><span class="secno">2.1 </span>Create a Project</h3>
<p> A registered user is to open the <a href="/new/">Add New Project</a> link:
</p><ul>
    <li> <span>Project name</span> should be unique among all project names; it is recommended to add the language pair suffix in
        the name like HTML5-enru for English =&gt; Russian pair.
    </li><li> Fill in <span>Source Language</span> and <span>Target Language</span> fields for your project language pair; 
        if you really want to choose <em>territory or country variety</em> of a language, then you'd rather note it in the project's name like
        HTML5-enpt_br for Brazilian Portuguese.
    </li><li> Fill in other optional fields and complete the project creation.
    </li><li> You are the <span>Project Owner</span> now.
</li></ul>
</body>`,
				},
			},
			Stream:      true,
			Temperature: 0, // 0.00001, //
			TopP:        0, // 0.00001, //
		}
		stream, err := client.CreateChatCompletionStream(context.Background(), req)
		if err != nil {
			fmt.Printf("ChatCompletionStream error: %v\n", err)
			return
		}
		defer stream.Close()

		fmt.Printf("Stream response: \n")
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
