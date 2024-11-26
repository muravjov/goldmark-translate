package markdown

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"

	utilerr "git.catbo.net/muravjov/go2023/util"
)

type nodeRenderer struct {
}

func NewNodeRenderer() NodeRenderer {
	r := &nodeRenderer{}

	return r
}

func (r *nodeRenderer) RegisterFuncs(reg NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)

	// inlines

	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)

}

func (r *nodeRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString(strings.Repeat("#", n.Level) + " ")
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderBlockquote(
	w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("> ")
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		w.Write(line.Value(source))
	}
}

func (r *nodeRenderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		_, _ = w.WriteString("```")
		language := n.Language(source)
		if language != nil {
			w.Write(language)
		}
		_, _ = w.WriteString("\n")
		r.writeLines(w, source, n)
	} else {
		_, _ = w.WriteString("```")
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
		prefix := "- "
		if list.IsOrdered() {
			prefix = fmt.Sprintf("%v. ", list.Start)
		}

		_, _ = w.WriteString(prefix)
	} else {
		if n.NextSibling() != nil {
			sep := "\n"
			if !list.IsTight {
				sep = "\n\n"
			}

			_, _ = w.WriteString(sep)
		}
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) renderEmphasis(
	w util.BufWriter, _ []byte, node ast.Node, _ bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)

	// * or _
	_, _ = w.WriteString(strings.Repeat("*", n.Level))
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

func (r *nodeRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	w.Write(segment.Value(source))
	return ast.WalkContinue, nil
}
