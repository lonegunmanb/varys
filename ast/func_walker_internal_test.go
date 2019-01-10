package ast

//go:generate mockgen -package=ast -destination=./mock_receiver_type_retriever.go github.com/lonegunmanb/varys/ast ReceiverTypeRetriever

import (
	"github.com/golang/mock/gomock"
	"github.com/lonegunmanb/johnnie"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"testing"
)

//TODO:finish this test after we finished all tests based on mock
//func TestVisitMethod(t *testing.T) {
//	code := `
//package ast
//type Struct struct {
//}
//
//func (s *Struct) Method() {
//}`
//	walker := parseCodeWithFuncWalker(t, code)
//	methods := walker.methods
//	assert.Equal(t, 1, len(methods))
//	assert.Equal(t, "Method", methods[0].GetName())
//}

func TestWalkFuncDecl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	name := "name"
	stubTypeInfo := &typeInfo{}
	stubExpr := &ast.StructType{}
	mockTypeRetriever := NewMockReceiverTypeRetriever(ctrl)
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(stubExpr)).Times(1).Return(stubTypeInfo)
	funcDecl := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{Type: stubExpr},
			},
		},
		Name: &ast.Ident{Name: name},
	}
	sut := NewFuncWalker(mockTypeRetriever)
	johnnie.Visit(sut, funcDecl)
	methods := sut.GetMethods()
	assert.Equal(t, 1, len(methods))
	method := methods[0]
	assert.Equal(t, name, method.GetName())
	assert.Equal(t, stubTypeInfo, method.GetReceiver())
}
