package markdown

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/util"
)

type Context struct {
	PaddingStack []string

	counter int
	verbose bool
}

func NewContext(verbose bool) *Context {
	r := &Context{
		verbose: verbose,
	}

	return r
}

func (c *Context) PushStack(pad string) {
	c.PaddingStack = append(c.PaddingStack, pad)
}

func (c *Context) PopStack() {
	c.PaddingStack = c.PaddingStack[:len(c.PaddingStack)-1]
}

func (c *Context) Pad(w util.BufWriter) {
	if c.verbose {
		_, _ = w.WriteString(fmt.Sprintf("%v", c.counter))
		c.counter++
	}

	if stack := c.PaddingStack; len(stack) > 0 {
		_, _ = w.WriteString(strings.Join(stack, ""))
	}
}
