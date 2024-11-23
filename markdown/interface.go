package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

type NodeRendererFunc func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error)

// A NodeRenderer interface offers NodeRendererFuncs.
type NodeRenderer interface {
	// RendererFuncs registers NodeRendererFuncs to given NodeRendererFuncRegisterer.
	RegisterFuncs(NodeRendererFuncRegisterer)
}

// A NodeRendererFuncRegisterer registers given NodeRendererFunc to this object.
type NodeRendererFuncRegisterer interface {
	// Register registers given NodeRendererFunc to this object.
	Register(ast.NodeKind, NodeRendererFunc)
}
