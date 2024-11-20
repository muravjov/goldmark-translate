package markdown

import (
	"io"

	"github.com/yuin/goldmark/ast"
	goldrender "github.com/yuin/goldmark/renderer"
)

type Renderer struct {
}

func NewRenderer() goldrender.Renderer {
	r := &Renderer{
		//config:               config,
		//nodeRendererFuncsTmp: map[ast.NodeKind]NodeRendererFunc{},
	}

	return r
}

func (r *Renderer) AddOptions(_ ...goldrender.Option) {
	// Nothing to add
}

func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	return nil
}
