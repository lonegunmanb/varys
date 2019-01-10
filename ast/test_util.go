package ast

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testPkgPath = "github.com/lonegunmanb/varys/ast"
const testPhysicalPath = "Users/test/go/src/lonegunmanb/varys/ast"

func parseCode(t *testing.T, sourceCode string) *typeWalker {
	typeWalker := NewTypeWalker().(*typeWalker)
	typeWalker.physicalPath = testPhysicalPath
	ctrl := gomock.NewController(t)
	mockOsEnv := NewMockGoPathEnv(ctrl)
	mockOsEnv.EXPECT().GetPkgPath(gomock.Eq(testPhysicalPath)).Times(1).Return(testPkgPath, nil)
	typeWalker.osEnv = mockOsEnv
	err := typeWalker.Parse(testPkgPath, sourceCode)
	assert.Nil(t, err)
	return typeWalker
}
