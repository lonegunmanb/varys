package ast

import (
	"github.com/ahmetb/go-linq"
	"github.com/lonegunmanb/johnnie"
	"go/ast"
	"go/types"
)

type FuncWalker interface {
	johnnie.Walker
	GetMethods() []MethodInfo
}

type funcWalker struct {
	AbstractWalker
	methods             []MethodInfo
	structTypeRetriever TypeRetriever
}

func (walker *funcWalker) GetMethods() []MethodInfo {
	return walker.methods
}

func NewFuncWalker(typeRetriever TypeRetriever) FuncWalker {
	if typeRetriever == nil {
		panic("no struct type info")
	}
	walker := &funcWalker{
		structTypeRetriever: typeRetriever,
	}
	walker.AbstractWalker = *newAbstractWalker(walker)
	return walker
}

func (walker *funcWalker) WalkFuncDecl(funcDecl *ast.FuncDecl) bool {
	isMethod := funcDecl.Recv != nil && len(funcDecl.Recv.List) == 1
	if isMethod {
		funcType := funcDecl.Type
		methodInfo := &methodInfo{
			name:        funcDecl.Name.Name,
			receiver:    walker.getTypeInfo(funcDecl.Recv.List[0].Type),
			returnTypes: walker.analyzeReturnTypes(funcType),
		}
		walker.methods = append(walker.methods, methodInfo)
	}
	return false
}

func (walker *funcWalker) analyzeReturnTypes(funcType *ast.FuncType) []types.Type {
	returnTypes := make([]types.Type, 0, len(funcType.Results.List))
	linq.From(funcType.Results.List).SelectMany(func(field interface{}) linq.Query {
		f := field.(*ast.Field)
		fieldType := walker.getType(f.Type)
		if len(f.Names) == 0 {
			return linq.Repeat(fieldType, 1)
		}
		return linq.From(f.Names).Select(func(_ interface{}) interface{} {
			return fieldType
		})
	}).ToSlice(&returnTypes)
	return returnTypes
}

func (walker *funcWalker) getType(expr ast.Expr) types.Type {
	return walker.structTypeRetriever.GetType(expr)
}

func (walker *funcWalker) getTypeInfo(expr ast.Expr) TypeInfo {
	return walker.structTypeRetriever.GetTypeInfo(expr)
}
