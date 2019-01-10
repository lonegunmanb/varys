package ast

import (
	"go/ast"
)

type funcWalker struct {
	AbstractWalker
	methods []MethodInfo
}

func newFuncWalker() *funcWalker {
	walker := &funcWalker{}
	walker.AbstractWalker = *newAbstractWalker(walker)
	return walker
}

func (walker *funcWalker) WalkFuncDecl(d *ast.FuncDecl) {
	isMethod := d.Recv != nil && len(d.Recv.List) == 1
	if isMethod {
		methodInfo := &methodInfo{name: d.Name.Name}
		walker.methods = append(walker.methods, methodInfo)
	}
}
