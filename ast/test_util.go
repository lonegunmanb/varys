package ast

//go:generate mockgen -source=./gopath.go -package=ast -destination=./mock_gopathenv.go

import (
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testPkgPath = "github.com/lonegunmanb/varys/ast"
const testPhysicalPath = "Users/test/go/src/lonegunmanb/varys/ast"

var Given = convey.Convey
var When = convey.Convey
var Then = convey.Convey
var And = convey.So

func prepareTypeWalker(t *testing.T) *typeWalker {
	walker := NewTypeWalker().(*typeWalker)
	walker.physicalPath = testPhysicalPath
	ctrl := gomock.NewController(t)
	mockOsEnv := NewMockGoPathEnv(ctrl)
	mockOsEnv.EXPECT().GetPkgPath(gomock.Eq(testPhysicalPath)).AnyTimes().Return(testPkgPath, nil)
	walker.osEnv = mockOsEnv
	return walker
}

func assertSame(t *testing.T, p1 interface{}, p2 interface{}) {
	assert.True(t, p1 == p2)
}

//noinspection GoUnusedFunction
func assertNotSame(t *testing.T, p1 interface{}, p2 interface{}) {
	assert.False(t, p1 == p2)
}
