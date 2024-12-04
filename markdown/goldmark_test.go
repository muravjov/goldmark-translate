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
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
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

	var data []byte
	if true {

		if false {
			// :TODO: improper rendering

			data = []byte(`
> Lorem ipsum dolor
sit amet.
> - Qui *quodsi iracundia*
> - aliquando id
`)

		}

		if false {
			// :TODO: improper rendering

			// https://spec.commonmark.org/dingus/?text=1.%20a%0A%0A%20%202.%20b%0A%0A%20%20%20%203.%20c%0A
			// - commonmark.js parses para-s in list items into para-s
			// - goldmark parses para-s in list itemes into TextBlock-s
			// - both parse other types like fenced code blocks into fenced code blocks
			data = []byte(`
 - Qui *quodsi iracundia*
 - aliquando id
   - hello
 - ~~~
   fdsfds
   ~~~
`)

		}

		if false {
			data = []byte(`
2. ggg

2. ddd
2. ads
`)
		}

		if false {
			// :TODO: blockquotes with non trivial blocks (not a para or multiple blocks)
			// is rendered improperly: intendation with "> " needed
			data = []byte(`
> fdsa
> fds
> fdsa
>
> fds
`)
		}

		if false {
			// SoftLineBreak
			data = []byte(`
This document is a detailed reference manual for Go's module system. For an
introduction to creating Go projects, see [How to Write Go
Code](/doc/code.html). For information on using modules,
migrating projects to modules, and other topics, see the blog series starting
with [Using Go Modules](/blog/using-go-modules).`)
		}

		if false {
			// HardLineBreak
			data = []byte(`
foo  
baz

foo\
baz

*foo  
bar*`)
		}

		if false {
			// attributes:
			// - https://github.com/yuin/goldmark#attributes
			// - https://talk.commonmark.org/t/consistent-attribute-syntax/272
			data = []byte(`
## Introduction {#introduction}

Modules are how Go manages dependencies.
`)
		}

		if false {
			// CodeSpan like `go.mod`
			// and
			// RawHTML like <dfn>
			data = []byte(`
A module is identified by a [module path](#glos-module-path), which is declared
in a [` + "`" + `go.mod` + "`" + ` file](#go-mod-file), together with information about the
module's dependencies. The <dfn>module root directory</dfn> is the directory
that contains the ` + "`" + `go.mod` + "`" + ` file. The <dfn>main module</dfn> is the module
containing the directory where the ` + "`" + `go` + "`" + ` command is invoked.
`)
		}

		if false {
			// nested blocks (in lists and blockquotes)
			data = []byte(`

* item 1

  para 1.1
  para 1.1 continuation

  para 1.2

  > bq 1
  > bq para 1
  
  * subitem 1

    para 1.2.1

  * subitem 2

    para 1.2.2

para 2

* tight item 3
* tight item 4

`)
		}

		if false {
			// list custom offsets
			data = []byte(`

* At ` + "`" + `go 1.21` + "`" + ` or higher:
   * The ` + "`" + `go` + "`" + ` line declares a required minimum version of Go to use with this module.
   * The ` + "`" + `go` + "`" + ` line must be greater than or equal to the ` + "`" + `go` + "`" + ` line of all dependencies.
   * The ` + "`" + `go` + "`" + ` command no longer attempts to maintain compatibility with the previous older version of Go.
   * The ` + "`" + `go` + "`" + ` command is more careful about keeping checksums of ` + "`" + `go.mod` + "`" + ` files in the ` + "`" + `go.sum` + "`" + ` file.
`)
		}

		if false {
			// list breaks paragraph
			data = []byte(`
The number of windows in my house is

11. The number of doors is 6.

The number of windows in my house is
1. The number of doors is 6.

The number of windows in my house is
* The number of doors is 6.

The number of windows in my house is

* The number of doors is 6.
`)
		}

		if true {
			// raw text with end of lines
			data = []byte(`
* At ` + "`" + `go 1.17` + "`" + ` or higher:
   * The ` + "`" + `go.mod` + "`" + ` file includes an explicit [` + "`" + `require` + "`" + `
     directive](#go-mod-file-require) for each module that provides any package
     transitively imported by a package or test in the main module. (At ` + "`" + `go
     1.16` + "`" + ` and lower, an [indirect dependency](#glos-direct-dependency) is
     included only if [minimal version selection](#minimal-version-selection)
     would otherwise select a different version.) This extra information enables
     [module graph pruning](#graph-pruning) and [lazy module
     loading](#lazy-loading).
`)
		}

	} else {
		var err error
		fName := "/Users/ilya/opt/programming/catbo/stuff/medium/Monitoring1.md"
		data, err = os.ReadFile(fName)
		assert.NoError(t, err)
	}

	md2md := true           // false //
	verbosePadding := false // true //

	var writer io.Writer = os.Stderr // os.Stdout //
	writer = ZeroBufWriter{writer}

	err := Convert(data, writer, md2md, true, verbosePadding)
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
