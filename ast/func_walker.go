package ast

import (
	"github.com/lonegunmanb/johnnie"
	"go/ast"
)

type FuncWalker interface {
	johnnie.Walker
	GetMethods() []MethodInfo
}

type funcWalker struct {
	AbstractWalker
	methods             []MethodInfo
	structTypeRetriever ReceiverTypeRetriever
}

func (walker *funcWalker) GetMethods() []MethodInfo {
	return walker.methods
}

func NewFuncWalker(structTypeRetriever ReceiverTypeRetriever) FuncWalker {
	if structTypeRetriever == nil {
		panic("no struct type info")
	}
	walker := &funcWalker{
		structTypeRetriever: structTypeRetriever,
	}
	walker.AbstractWalker = *newAbstractWalker(walker)
	return walker
}

func (walker *funcWalker) WalkFuncDecl(funcDecl *ast.FuncDecl) bool {
	isMethod := funcDecl.Recv != nil && len(funcDecl.Recv.List) == 1
	if isMethod {
		methodInfo := &methodInfo{
			name:     funcDecl.Name.Name,
			receiver: walker.structTypeRetriever.GetType(funcDecl.Recv.List[0].Type),
		}
		walker.methods = append(walker.methods, methodInfo)
	}
	return false
}
