package ast

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testPkgPath = "github.com/lonegunmanb/varys/ast"
const testPhysicalPath = "Users/test/go/src/lonegunmanb/varys/ast"

func parseCodeWithTypeWalker(t *testing.T, sourceCode string) *typeWalker {
	typeWalker := NewTypeWalker().(*typeWalker)
	parseCode(t, sourceCode, &typeWalker.AbstractWalker)
	return typeWalker
}

//func parseCodeWithFuncWalker(t *testing.T, sourceCode string, retriever TypeRetriever) *funcWalker {
//	funcWalker := NewFuncWalker(retriever).(*funcWalker)
//	parseCode(t, sourceCode, &funcWalker.AbstractWalker)
//	return funcWalker
//}

func parseCode(t *testing.T, sourceCode string, walker *AbstractWalker) {
	walker.physicalPath = testPhysicalPath
	ctrl := gomock.NewController(t)
	mockOsEnv := NewMockGoPathEnv(ctrl)
	mockOsEnv.EXPECT().GetPkgPath(gomock.Eq(testPhysicalPath)).Times(1).Return(testPkgPath, nil)
	walker.osEnv = mockOsEnv
	err := walker.Parse(testPkgPath, sourceCode)
	assert.Nil(t, err)
}

//noinspection GoUnusedFunction
func assertSame(t *testing.T, p1 interface{}, p2 interface{}) {
	assert.True(t, p1 == p2)
}

//noinspection GoUnusedFunction
func assertNotSame(t *testing.T, p1 interface{}, p2 interface{}) {
	assert.False(t, p1 == p2)
}
