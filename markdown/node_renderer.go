package markdown

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type NodeRenderer struct {
}

func NewNodeRenderer() renderer.NodeRenderer {
	r := &NodeRenderer{}

	return r
}

func (r *NodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindHeading, r.renderHeading)

	// inlines

	reg.Register(ast.KindText, r.renderText)

}

func (r *NodeRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString(strings.Repeat("#", n.Level) + " ")
	} else {
		_, _ = w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// :TEMP!!!:
func RenderAttributes(w util.BufWriter, node ast.Node) {
	for _, attr := range node.Attributes() {
		_, _ = w.WriteString(" ")
		_, _ = w.Write(attr.Name)
		_, _ = w.WriteString(`="`)
		// TODO: convert numeric values to strings
		var value []byte
		switch typed := attr.Value.(type) {
		case []byte:
			value = typed
		case string:
			value = util.StringToReadOnlyBytes(typed)
		}
		_, _ = w.Write(util.EscapeHTML(value))
		_ = w.WriteByte('"')
	}
}

func (r *NodeRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	w.Write(segment.Value(source))
	return ast.WalkContinue, nil
}
