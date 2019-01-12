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
	stubParameterTypeExpr := &ast.BasicLit{}
	stubReturnTypeExpr := &ast.InterfaceType{}
	expectedPameterType := &types.Basic{}
	expectedReturnType := &types.Interface{}
	mockTypeRetriever := NewMockTypeRetriever(ctrl)
	mockTypeRetriever.EXPECT().GetTypeInfo(gomock.Eq(stubReceiverTypeExpr)).Times(1).Return(stubTypeInfo)
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(stubReturnTypeExpr)).Times(1).Return(expectedReturnType)
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(stubParameterTypeExpr)).Times(1).Return(expectedPameterType)
	funcDecl := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{Type: stubReceiverTypeExpr},
			},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Type: stubParameterTypeExpr},
				},
			},
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
	parameterFields := method.GetParameterFields()
	parameterField := parameterFields[0]
	assertSame(t, expectedPameterType, parameterField.GetType())
	assert.Equal(t, 1, len(parameterFields))
	returnTypes := method.GetReturnFields()
	assert.Equal(t, 1, len(returnTypes))
	returnType := returnTypes[0]
	assertSame(t, expectedReturnType, returnType.GetType())
}

func TestMultipleNameFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTypeRetriever := NewMockTypeRetriever(ctrl)
	stubTypeExpr := &ast.InterfaceType{}
	expectedType := &types.Interface{}
	typeInfo := &typeInfo{}
	methodInfo := &methodInfo{
		receiver: typeInfo,
	}
	parameterName1 := "parameterName1"
	parameterName2 := "parameterName2"
	returnFieldName1 := "returnFieldName1"
	returnFieldName2 := "returnFieldName2"
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(stubTypeExpr)).Times(2).Return(expectedType)
	funcType := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: stubTypeExpr,
					Names: []*ast.Ident{
						{Name: parameterName1},
						{Name: parameterName2},
					},
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: stubTypeExpr,
					Names: []*ast.Ident{
						{Name: returnFieldName1},
						{Name: returnFieldName2},
					},
				},
			},
		},
	}
	sut := NewFuncWalker(mockTypeRetriever).(*funcWalker)
	parameterTypes := sut.analyzeTypes(methodInfo, funcType.Params)
	returnTypes := sut.analyzeTypes(methodInfo, funcType.Results)

	assertFuncFields(t, returnTypes, expectedType, returnFieldName1, returnFieldName2, methodInfo, typeInfo)
	assertFuncFields(t, parameterTypes, expectedType, parameterName1, parameterName2, methodInfo, typeInfo)
}

func assertFuncFields(t *testing.T, fieldInfos []FieldInfo, expectedType *types.Interface, fieldName1 string, fieldName2 string, methodInfo *methodInfo, typeInfo *typeInfo) {
	assert.Equal(t, 2, len(fieldInfos))
	assertSame(t, expectedType, fieldInfos[0].GetType())
	assert.Equal(t, fieldName1, fieldInfos[0].GetName())
	assertSame(t, methodInfo, fieldInfos[0].GetReferenceFromMethod())
	assertSame(t, typeInfo, fieldInfos[0].GetReferenceFromType())
	assertSame(t, expectedType, fieldInfos[1].GetType())
	assert.Equal(t, fieldName2, fieldInfos[1].GetName())
	assertSame(t, methodInfo, fieldInfos[1].GetReferenceFromMethod())
	assertSame(t, typeInfo, fieldInfos[1].GetReferenceFromType())
}
