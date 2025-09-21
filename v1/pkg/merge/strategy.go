package merge

import (
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// MergeStrategy defines the merge behavior interface
type MergeStrategy interface {
	Name() string
	Merge(base, override node.Node, ctx *Context) (node.Node, error)
}
