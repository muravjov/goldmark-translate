package markdown

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// COPY_N_PASTE of golang.org/x/website

func markdownToHTMLWithDump(markdown string, myDumpFile func(filenameSuffix string, data string)) (template.HTML, error) {
	// parser.WithHeadingAttribute allows custom ids on headings.
	// html.WithUnsafe allows use of raw HTML, which we need for tables.
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithHeadingAttribute(),
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(util.Prioritized(mdTransformFunc(mdLink), 1)),
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
		goldmark.WithExtensions(
			extension.NewTypographer(),
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols([][]byte{[]byte("http"), []byte("https")}),
				extension.WithLinkifyEmailRegexp(regexp.MustCompile(`[^\x00-\x{10FFFF}]`)), // impossible
			),
			extension.DefinitionList,
			extension.NewTable(),
		),
	)

	data := replaceTabs([]byte(markdown))
	myDumpFile(".catbo.md.src.txt", string(data))

	var buf bytes.Buffer
	if err := md.Convert(data, &buf); err != nil {
		return "", err
	}

	myDumpFile(".catbo.md.dst.html", string(buf.Bytes()))

	return template.HTML(buf.Bytes()), nil
}

// mdTransformFunc is a func implementing parser.ASTTransformer.
type mdTransformFunc func(*ast.Document, text.Reader, parser.Context)

func (f mdTransformFunc) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	f(node, reader, pc)
}

// mdLink walks doc, adding rel=noreferrer target=_blank to non-relative links.
func mdLink(doc *ast.Document, _ text.Reader, _ parser.Context) {
	mdLinkWalk(doc)
}

func mdLinkWalk(n ast.Node) {
	switch n := n.(type) {
	case *ast.Link:
		dest := string(n.Destination)
		if strings.HasPrefix(dest, "https://") || strings.HasPrefix(dest, "http://") {
			n.SetAttributeString("rel", []byte("noreferrer"))
			n.SetAttributeString("target", []byte("_blank"))
		}
		return
	case *ast.AutoLink:
		// All autolinks are non-relative.
		n.SetAttributeString("rel", []byte("noreferrer"))
		n.SetAttributeString("target", []byte("_blank"))
		return
	}

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		mdLinkWalk(child)
	}
}

// replaceTabs replaces all tabs in text with spaces up to a 4-space tab stop.
//
// In Markdown, tabs used for indentation are required to be interpreted as
// 4-space tab stops. See https://spec.commonmark.org/0.30/#tabs.
// Go also renders nicely and more compactly on the screen with 4-space
// tab stops, while browsers often use 8-space.
// And Goldmark crashes in some inputs that mix spaces and tabs.
// Fix the crashes and make the Go code consistently compact across browsers,
// all while staying Markdown-compatible, by expanding to 4-space tab stops.
//
// This function does not handle multi-codepoint Unicode sequences correctly.
func replaceTabs(text []byte) []byte {
	var buf bytes.Buffer
	col := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		text = text[size:]

		switch r {
		case '\n':
			buf.WriteByte('\n')
			col = 0

		case '\t':
			buf.WriteByte(' ')
			col++
			for col%4 != 0 {
				buf.WriteByte(' ')
				col++
			}

		default:
			buf.WriteRune(r)
			col++
		}
	}
	return buf.Bytes()
}

// COPY_N_PASTE of golang.org/x/website: end

func TestGolangWebsiteMD(t *testing.T) {
	t.SkipNow()

	myLogDir := "/Users/ilya/opt/programming/golang/tmp/golang-website-logs"
	filenamePrefix := strings.ReplaceAll("/ref/mod", "/", "-")
	myDumpFile := func(filenameSuffix string, data string) {
		filename := fmt.Sprintf("%v/%v%v", myLogDir, filenamePrefix, filenameSuffix)
		fmt.Fprintf(os.Stderr, "Writing %v ...\n", filename)
		if err := os.WriteFile(filename, []byte(data), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "os.WriteFile error: %v", err)
		}
	}

	dat, err := os.ReadFile(fmt.Sprintf("%v/%v%v", myLogDir, filenamePrefix, ".executed.src.txt"))
	assert.NoError(t, err)

	_, err = markdownToHTMLWithDump(string(dat), myDumpFile)
	assert.NoError(t, err)
}

func TestMarkdown2Markdown(t *testing.T) {
	//t.SkipNow()

	fName := "/Users/ilya/opt/programming/catbo/stuff/medium/Monitoring1.md"
	data, err := os.ReadFile(fName)
	assert.NoError(t, err)

	md2md := true // false //

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
	writer := os.Stderr // os.Stdout //
	source := data

	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	doc.Dump(reader.Source(), 0)

	err = md.Renderer().Render(ZeroBufWriter{writer}, source, doc)
	assert.NoError(t, err)
}

// no way to implement interface on another interface
//type ZeroBufWriter io.Writer

type ZeroBufWriter struct {
	io.Writer
}

func (w ZeroBufWriter) Available() int {
	return 0
}

func (w ZeroBufWriter) Buffered() int {
	return 0
}

func (w ZeroBufWriter) Flush() error {
	return nil
}

func (w ZeroBufWriter) WriteByte(c byte) error {
	_, err := w.Write([]byte{c})
	return err
}

func (w ZeroBufWriter) WriteRune(r rune) (int, error) {
	return w.WriteString(string(r))
}

func (w ZeroBufWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
