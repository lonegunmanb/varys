package ast

//go:generate mockgen -package=ast -destination=./mock_type_retriever.go github.com/lonegunmanb/varys/ast TypeRetriever

import (
	"github.com/golang/mock/gomock"
	"github.com/lonegunmanb/johnnie"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/types"
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
	stubReceiverTypeExpr := &ast.StructType{}
	stubReturnTypeExpr := &ast.InterfaceType{}
	expectedReturnType := &types.Interface{}
	mockTypeRetriever := NewMockTypeRetriever(ctrl)
	mockTypeRetriever.EXPECT().GetTypeInfo(gomock.Eq(stubReceiverTypeExpr)).Times(1).Return(stubTypeInfo)
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(stubReturnTypeExpr)).Times(1).Return(expectedReturnType)
	funcDecl := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{Type: stubReceiverTypeExpr},
			},
		},
		Type: &ast.FuncType{
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: stubReturnTypeExpr},
				},
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
	assertSame(t, stubTypeInfo, method.GetReceiver())
	returnTypes := method.GetReturnFields()
	assert.Equal(t, 1, len(method.GetReturnFields()))
	returnType := returnTypes[0]
	assertSame(t, expectedReturnType, returnType.GetType())
}

func TestMultipleNameReturnField(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTypeRetriever := NewMockTypeRetriever(ctrl)
	stubReturnTypeExpr := &ast.InterfaceType{}
	expectedReturnType := &types.Interface{}
	typeInfo := &typeInfo{}
	methodInfo := &methodInfo{
		receiver: typeInfo,
	}
	name1 := "name1"
	name2 := "name2"
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(stubReturnTypeExpr)).Times(1).Return(expectedReturnType)
	returnType := &ast.FuncType{
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: stubReturnTypeExpr,
					Names: []*ast.Ident{
						{Name: name1},
						{Name: name2},
					},
				},
			},
		},
	}
	sut := NewFuncWalker(mockTypeRetriever).(*funcWalker)
	returnTypes := sut.analyzeReturnTypes(returnType, methodInfo)

	assert.Equal(t, 2, len(returnTypes))
	assertSame(t, expectedReturnType, returnTypes[0].GetType())
	assert.Equal(t, name1, returnTypes[0].GetName())
	assertSame(t, methodInfo, returnTypes[0].GetReferenceFromMethod())
	assertSame(t, typeInfo, returnTypes[0].GetReferenceFromType())
	assertSame(t, expectedReturnType, returnTypes[1].GetType())
	assert.Equal(t, name2, returnTypes[1].GetName())
	assertSame(t, methodInfo, returnTypes[1].GetReferenceFromMethod())
	assertSame(t, typeInfo, returnTypes[1].GetReferenceFromType())
}
