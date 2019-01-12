package ast

import (
	"go/ast"
	"go/types"
)

type TypeRetriever interface {
	GetTypeInfo(expr ast.Expr) TypeInfo
	GetType(expr ast.Expr) types.Type
}
