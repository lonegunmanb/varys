package ast

//go:generate mockgen -source=./type_retriever.go -package=ast -destination=./mock_type_retriever.go

import (
	"github.com/golang/mock/gomock"
	"github.com/lonegunmanb/johnnie"
	. "github.com/smartystreets/goconvey/convey"
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

type walkFuncDeclTestData struct {
	name                  string
	typeInfo              *typeInfo
	receiverTypeExpr      *ast.StructType
	parameterTypeExpr     *ast.BasicLit
	returnTypeExpr        *ast.InterfaceType
	expectedParameterType *types.Basic
	expectedReturnType    *types.Interface
}

func TestWalkFuncDecl(t *testing.T) {
	Convey("given func decl", t, func() {
		testData := &walkFuncDeclTestData{
			name:                  "name",
			typeInfo:              &typeInfo{},
			receiverTypeExpr:      &ast.StructType{},
			parameterTypeExpr:     &ast.BasicLit{},
			returnTypeExpr:        &ast.InterfaceType{},
			expectedParameterType: &types.Basic{},
			expectedReturnType:    &types.Interface{},
		}
		ctrl, mockTypeRetriever := setupMockTypeRetriever(t, testData)
		defer ctrl.Finish()
		funcDecl := createFuncDecl(testData)
		sut := NewFuncWalker(mockTypeRetriever)
		Convey("when visit func decl with func walker", func() {
			johnnie.Visit(sut, funcDecl)
			Convey("then we gather correct methodInfo", func() {
				So(sut, shouldGatherExpectedMethodInfo, testData)
			})
		})
	})
}

func shouldGatherExpectedMethodInfo(actual interface{}, expected ...interface{}) string {
	sut := actual.(FuncWalker)
	testData := expected[0].(*walkFuncDeclTestData)
	methods := sut.GetMethods()
	So(len(methods), ShouldEqual, 1)
	method := methods[0]
	So(method.GetName(), ShouldEqual, testData.name)
	So(method.GetReceiver(), shouldBeSame, testData.typeInfo)
	parameterFields := method.GetParameterFields()
	parameterField := parameterFields[0]
	So(len(parameterFields), ShouldEqual, 1)
	So(parameterField.GetType(), ShouldEqual, testData.expectedParameterType)
	returnTypes := method.GetReturnFields()
	So(len(returnTypes), ShouldEqual, 1)
	So(returnTypes[0].GetType(), ShouldEqual, testData.expectedReturnType)
	return ""
}

func shouldBeSame(actual interface{}, expected ...interface{}) string {
	if len(expected) == 1 && actual == expected[0] {
		return ""
	}
	return "not same pointer"
}

func createFuncDecl(testData *walkFuncDeclTestData) *ast.FuncDecl {
	funcDecl := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{Type: testData.receiverTypeExpr},
			},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Type: testData.parameterTypeExpr},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: testData.returnTypeExpr},
				},
			},
		},
		Name: &ast.Ident{Name: testData.name},
	}
	return funcDecl
}

func setupMockTypeRetriever(t *testing.T, testData *walkFuncDeclTestData) (*gomock.Controller, *MockTypeRetriever) {
	ctrl := gomock.NewController(t)
	mockTypeRetriever := NewMockTypeRetriever(ctrl)
	mockTypeRetriever.EXPECT().GetTypeInfo(gomock.Eq(testData.receiverTypeExpr)).Times(1).Return(testData.typeInfo)
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(testData.returnTypeExpr)).Times(1).Return(testData.expectedReturnType)
	mockTypeRetriever.EXPECT().GetType(gomock.Eq(testData.parameterTypeExpr)).Times(1).Return(testData.expectedParameterType)
	return ctrl, mockTypeRetriever
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
