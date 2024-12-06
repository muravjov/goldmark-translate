package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"

	utilerr "git.catbo.net/muravjov/go2023/util"
)

type nodeRenderer struct {
	context *Context
}

func NewNodeRenderer(context *Context) NodeRenderer {
	r := &nodeRenderer{
		context: context,
	}

	return r
}

func (r *nodeRenderer) RegisterFuncs(reg NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// inlines

	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)

}

var attrNameID = []byte("id")
var attrNameClass = []byte("class")

func (r *nodeRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString(strings.Repeat("#", n.Level) + " ")
	} else {
		if node.Attributes() != nil {
			_, _ = w.WriteString(" {")
			first := true
			for _, attr := range node.Attributes() {
				if !first {
					w.WriteByte(' ')
				}
				first = false

				if bytes.Equal(attr.Name, attrNameID) {
					w.WriteByte('#')
				} else if bytes.Equal(attr.Name, attrNameClass) {
					w.WriteByte('.')
				} else {
					_, _ = w.Write(attr.Name)
					w.WriteByte('=')
				}

				var value []byte
				switch typed := attr.Value.(type) {
				case []byte:
					value = typed
				case string:
					value = util.StringToReadOnlyBytes(typed)
				}
				_, _ = w.Write(value)
			}
			_, _ = w.WriteString("}")
		}
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderBlockquote(
	w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("> ")

		r.context.PushStack("> ")
	} else {
		r.context.PopStack()
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) writeLines(w util.BufWriter, source []byte, n ast.Node, codeBlock bool) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		b := line.Value(source)
		if codeBlock && !bytes.Equal(b, []byte{'\n'}) {
			_, _ = w.WriteString("    ") // 4 spaces
		}
		_, _ = w.Write(b)
		if !(codeBlock && (i == l-1)) {
			r.context.Pad(w)
		}
	}
}

func (r *nodeRenderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	r.writeLines(w, source, n, true)

	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)

	fenceMarker := "```"
	lines := n.Lines()
	if l := lines.Len(); l > 0 {
		pos := lines.At(l - 1).Stop
		// quoting:
		// The closing code fence may be preceded by up to three spaces of indentation, and may be followed only by spaces or tabs, which are ignored.
		for i := pos; i < len(source); i++ {
			if source[i] == ' ' {
				continue
			}

			if i-pos > 3 {
				break
			}

			if source[i] != '~' {
				break
			}

			fenceMarker = "~~~"
		}
	}

	if entering {
		_, _ = w.WriteString(fenceMarker)
		language := n.Language(source)
		if language != nil {
			w.Write(language)
		}
		_, _ = w.WriteString("\n")
		r.context.Pad(w)

		r.writeLines(w, source, n, false)
	} else {
		_, _ = w.WriteString(fenceMarker)
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderHTMLBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			w.Write(line.Value(source))

			if n.HasClosure() || i != l-1 {
				r.context.Pad(w)
			}
		}
	} else {
		if n.HasClosure() {
			closure := n.ClosureLine
			w.Write(closure.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	//n := node.(*ast.List)
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	list, ok := n.Parent().(*ast.List)
	if !ok {
		utilerr.Errorf("expected list node but got: %v", n.Parent().Kind().String())
		return ast.WalkStop, nil
	}

	if entering {
		var prefix string
		if list.IsOrdered() {
			prefix = fmt.Sprintf("%v%c ", list.Start, list.Marker)
		} else {
			prefix = string(list.Marker) + " "
		}

		if diff := n.(*ast.ListItem).Offset - len(prefix); diff > 0 {
			// golang docs style: if offset more then just prefix, let's pad it with spaces at left
			prefix = strings.Repeat(" ", diff) + prefix
		}

		_, _ = w.WriteString(prefix)

		r.context.PushStack(strings.Repeat(" ", len(prefix)))
	} else {
		r.context.PopStack()
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderThematicBreak(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString("***\n")

	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.AutoLink)
	_, _ = w.Write(n.Label(source))
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	_ = w.WriteByte('`')
	r.renderTexts(w, source, n)
	_ = w.WriteByte('`')
	return ast.WalkSkipChildren, nil
}

func (r *nodeRenderer) renderEmphasis(
	w util.BufWriter, source []byte, node ast.Node, _ bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)

	// * or _
	marker := "*"

	// :TRICKY: better to support .Marker attribute in ast.Emphasis struct, saving it
	// in last successful emphasisDelimiterProcessor.CanOpenCloser() from opener.Char and
	// initialising it in emphasisDelimiterProcessor.OnMatch()
	if t, ok := n.FirstChild().(*ast.Text); ok {
		// * or _ should not be followed by whitespace, let's get the previous character
		if pos := t.Segment.Start - 1; pos >= 0 && pos < len(source) {
			marker = string(source[pos])
		}
	}

	_, _ = w.WriteString(strings.Repeat(marker, n.Level))
	return ast.WalkContinue, nil
}

func finishLink(w util.BufWriter, destination []byte, title []byte) {
	_, _ = w.WriteString("](")
	w.Write(destination)
	if title != nil {
		_, _ = w.WriteString(` "`)
		_, _ = w.Write(title)
		_ = w.WriteByte('"')
	}
	_ = w.WriteByte(')')
}

func (r *nodeRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString("![")
	r.renderTexts(w, source, n)
	finishLink(w, n.Destination, n.Title)
	return ast.WalkSkipChildren, nil
}

func (r *nodeRenderer) renderString(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	_, _ = w.Write(n.Value)
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderTexts(w util.BufWriter, source []byte, n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok {
			_, _ = r.renderString(w, source, s, true)
		} else if t, ok := c.(*ast.Text); ok {
			_, _ = r.renderText(w, source, t, true)
		} else {
			r.renderTexts(w, source, c)
		}
	}
}

func (r *nodeRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("[")
	} else {
		finishLink(w, n.Destination, n.Title)
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderRawHTML(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.RawHTML)
	l := n.Segments.Len()
	for i := 0; i < l; i++ {
		segment := n.Segments.At(i)
		_, _ = w.Write(segment.Value(source))
	}
	return ast.WalkSkipChildren, nil
}

func rawWrite(writer util.BufWriter, source []byte, context *Context) {
	n := 0
	l := len(source)
	for i := 0; i < l; i++ {
		if source[i] == '\n' {
			_, _ = writer.Write(source[i-n : i+1])
			context.Pad(writer)
			n = 0
			continue
		}
		n++
	}
	if n != 0 {
		_, _ = writer.Write(source[l-n:])
	}
}

func (r *nodeRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	b := n.Segment.Value(source)
	if n.IsRaw() {
		rawWrite(w, b, r.context)
	} else {
		w.Write(b)
	}

	if n.HardLineBreak() {
		_, _ = w.WriteString(`\
`)
		r.context.Pad(w)
	} else if n.SoftLineBreak() {
		_, _ = w.WriteString("\n")
		r.context.Pad(w)
	}

	return ast.WalkContinue, nil
}
