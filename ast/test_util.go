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

func parseCodeWithFuncWalker(t *testing.T, sourceCode string) *funcWalker {
	funcWalker := newFuncWalker()
	parseCode(t, sourceCode, &funcWalker.AbstractWalker)
	return funcWalker
}

func parseCode(t *testing.T, sourceCode string, walker *AbstractWalker) {
	walker.physicalPath = testPhysicalPath
	ctrl := gomock.NewController(t)
	mockOsEnv := NewMockGoPathEnv(ctrl)
	mockOsEnv.EXPECT().GetPkgPath(gomock.Eq(testPhysicalPath)).Times(1).Return(testPkgPath, nil)
	walker.osEnv = mockOsEnv
	err := walker.Parse(testPkgPath, sourceCode)
	assert.Nil(t, err)
}
