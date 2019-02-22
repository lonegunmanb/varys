package ast

import (
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

const expectedPkgName = "github.com/lonegunmanb/syringe"

func TestGetPkgPathFromSystemPathUsingGoPath(t *testing.T) {

	testGetPkgPathFromSystemPath(t, []string{
		"/Users/user/go",
	},
		"/Users/user/go/src/github.com/lonegunmanb/syringe",
		expectedPkgName)
	testGetPkgPathFromSystemPath(t, []string{
		"/Users/user2/go",
		"/Users/user/go",
	},
		"/Users/user/go/src/github.com/lonegunmanb/syringe",
		expectedPkgName)
	testGetPkgPathFromSystemPath(t, []string{
		"/Users/user/go",
	},
		"/Users/user/go/src",
		"")
}

func TestGetPkgPathInWindows(t *testing.T) {
	Convey("given windows env", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockEnv := NewMockGoPathEnv(ctrl)
		mockEnv.EXPECT().IsWindows().MinTimes(1).Return(true)
		mockEnv.EXPECT().GetGoPath().MinTimes(1).Return("c:\\go")
		Convey("when get pkg path from windows go src path", func() {
			pkgPath, err := getPkgPath(mockEnv, "c:\\go\\src\\github.com\\lonegunmanb\\syringe")
			Convey("pkg path should equal to expected", func() {
				So(err, ShouldBeNil)
				So(pkgPath, ShouldEqual, expectedPkgName)
			})
		})
	})
}

func TestConcatFileNameWithPath(t *testing.T) {
	path := concatFileNameWithPath(false, "/Users/user/go", "file")
	assert.Equal(t, "/Users/user/go/file", path)
	path = concatFileNameWithPath(true, "c:\\go", "file")
	assert.Equal(t, "c:\\go\\file", path)
}

func testGetPkgPathFromSystemPath(t *testing.T, goPaths []string, systemPath string, expected string) {
	pkgPath, err := getPkgPathFromSystemPathUsingGoPath(false, goPaths, systemPath)
	assert.Nil(t, err)
	assert.Equal(t, expected, pkgPath)
}
