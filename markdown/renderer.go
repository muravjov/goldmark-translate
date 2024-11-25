package markdown

import (
	"bufio"
	"io"
	"sync"

	"github.com/yuin/goldmark/ast"
	goldrender "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Renderer struct {
	NodeRenderers util.PrioritizedSlice

	nodeRendererFuncsTmp map[ast.NodeKind]NodeRendererFunc
	maxKind              int
	nodeRendererFuncs    []NodeRendererFunc

	initSync sync.Once
}

func NewRenderer(ps ...util.PrioritizedValue) goldrender.Renderer {
	r := &Renderer{
		NodeRenderers:        ps,
		nodeRendererFuncsTmp: map[ast.NodeKind]NodeRendererFunc{},
	}

	return r
}

func (r *Renderer) AddOptions(_ ...goldrender.Option) {
	// Nothing to add
}

func (r *Renderer) Register(kind ast.NodeKind, v NodeRendererFunc) {
	r.nodeRendererFuncsTmp[kind] = v
	if int(kind) > r.maxKind {
		r.maxKind = int(kind)
	}
}

func (r *Renderer) Render(w io.Writer, source []byte, root ast.Node) error {
	r.initSync.Do(func() {
		r.NodeRenderers.Sort()
		l := len(r.NodeRenderers)
		for i := l - 1; i >= 0; i-- {
			v := r.NodeRenderers[i]
			nr, _ := v.Value.(NodeRenderer)
			nr.RegisterFuncs(r)
		}
		r.nodeRendererFuncs = make([]NodeRendererFunc, r.maxKind+1)
		for kind, nr := range r.nodeRendererFuncsTmp {
			r.nodeRendererFuncs[kind] = nr
		}
		r.nodeRendererFuncsTmp = nil
	})
	writer, ok := w.(util.BufWriter)
	if !ok {
		writer = bufio.NewWriter(w)
	}
	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		s := ast.WalkStatus(ast.WalkContinue)

		var f NodeRendererFunc
		if k := n.Kind(); k <= ast.NodeKind(r.maxKind) {
			f = r.nodeRendererFuncs[k]
		}

		if f != nil {
			sF, err := f(writer, source, n, entering)
			if err != nil {
				return sF, err
			}

			s = sF
		}

		if !entering && n.Parent() == root {
			_, _ = writer.WriteString("\n\n")
		}

		return s, nil
	})
	if err != nil {
		return err
	}
	return writer.Flush()
}