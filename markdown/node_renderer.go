package markdown

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
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
