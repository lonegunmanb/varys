package ast

import "go/ast"

type ReceiverTypeRetriever interface {
	GetType(expr ast.Expr) TypeInfo
}
