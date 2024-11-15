package llmrequest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	//t.SkipNow()

	llmProvider := llmGigachat // llmOpenai //
	logRequests := true        // false //

	client, err := MakeClient(llmProvider, logRequests)
	assert.NoError(t, err)

	if false {
		resp, err := client.Client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: client.GetLLModel(true),
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
			Model:     client.GetLLModel(true),
			MaxTokens: 20,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Lorem ipsum",
				},
			},
			Stream: true,
		}
		stream, err := client.Client.CreateChatCompletionStream(context.Background(), req)
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
		html := `<body>
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
</body>`

		stream, err := HTML2Markdown(client, html)
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
