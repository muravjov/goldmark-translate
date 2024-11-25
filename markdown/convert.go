package markdown

import (
	"io"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func Convert(source []byte, writer io.Writer, md2md bool, dump bool) error {
	options := []goldmark.Option{
		goldmark.WithParserOptions(
			parser.WithHeadingAttribute(),
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			// we don't need extension.Typographer for md2md because we want to keep source text without substitutions
			// like don't => don’t
			//extension.NewTypographer(),
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols([][]byte{[]byte("http"), []byte("https")}),
				extension.WithLinkifyEmailRegexp(regexp.MustCompile(`[^\x00-\x{10FFFF}]`)), // impossible
			),
			extension.DefinitionList,
			extension.NewTable(),
		),
	}

	if md2md {
		options = append(options,
			// we must overwrite default renderer (which is to html)
			goldmark.WithRenderer(NewRenderer(util.Prioritized(
				NewNodeRenderer(),
				400,
			))),
		)
	} else {
		options = append(options, []goldmark.Option{
			goldmark.WithParserOptions(
				parser.WithASTTransformers(util.Prioritized(mdTransformFunc(mdLink), 1)),
			),
			goldmark.WithRendererOptions(html.WithUnsafe()),
			goldmark.WithExtensions(
				extension.NewTypographer(),
			),
		}...)
	}

	md := goldmark.New(options...)

	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	if dump {
		doc.Dump(reader.Source(), 0)
	}

	return md.Renderer().Render(writer, source, doc)
}
