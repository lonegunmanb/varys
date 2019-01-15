package ast

import (
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
	"testing"
)

type FieldInfoInternalTestSuite struct {
	suite.Suite
	walker *typeWalker
}

func TestFieldInfoInternalTestSuite(t *testing.T) {
	suite.Run(t, new(FieldInfoInternalTestSuite))
}

func (suite *FieldInfoInternalTestSuite) SetupTest() {
	suite.walker = NewTypeWalker().(*typeWalker)
	suite.walker.physicalPath = testPhysicalPath
	ctrl := gomock.NewController(suite.T())
	mockOsEnv := NewMockGoPathEnv(ctrl)
	mockOsEnv.EXPECT().GetPkgPath(gomock.Eq(testPhysicalPath)).AnyTimes().Return(testPkgPath, nil)
	suite.walker.osEnv = mockOsEnv
}

func (suite *FieldInfoInternalTestSuite) TestGetNamedTypeStructFieldPkgPath() {
	Convey("given struct with named type struct field", suite.T(), func() {
		sourceCode := `
package ast
import (
"go/scanner"
"go/token"
)
type Struct struct {
	Err scanner.Error
	FileSet token.FileSet
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("the struct field dep pkg path should be right", func() {
				So(err, ShouldBeNil)
				structInfo := suite.walker.Types()[0]
				errField := structInfo.Fields[0]
				fileSetField := structInfo.Fields[1]
				So(lengthOfDepPkgPaths(errField), ShouldEqual, 1)
				So(errField.GetDepPkgPaths()[0], ShouldEqual, "go/scanner")
				So(lengthOfDepPkgPaths(fileSetField), ShouldEqual, 1)
				So(fileSetField.GetDepPkgPaths()[0], ShouldEqual, "go/token")
			})

		})
	})
}

func lengthOfDepPkgPaths(fieldInfo FieldInfo) int {
	return len(fieldInfo.GetDepPkgPaths())
}

func (suite *FieldInfoInternalTestSuite) TestStructFieldPkgPath() {
	table := []struct {
		given      string
		sourceCode string
		pkgPaths   []string
	}{
		{
			given: "given pointer to named type",
			sourceCode: `
package ast
import (
"go/scanner"
)
type Struct struct {
	Err *scanner.Error
}
`,
			pkgPaths: []string{"go/scanner"}},
		{
			given: "given slice of named type",
			sourceCode: `
package ast
import (
"go/types"
)
type Struct struct {
	_types []types.Type
}
`,
			pkgPaths: []string{"go/types"},
		},
		{
			given: "given array of named type",
			sourceCode: `
package ast
import (
"go/scanner"
)
type Struct struct {
	Err [1]scanner.Error
}
`,
			pkgPaths: []string{"go/scanner"},
		},
		{
			given: "given slice of pointer to named type",
			sourceCode: `
package ast
import (
"go/scanner"
)
type Struct struct {
	Err []*scanner.Error
}
`,
			pkgPaths: []string{"go/scanner"},
		},
		{
			given: "given map from named type to another named type",
			sourceCode: `
package ast
import (
"go/scanner"
"go/token"
)
type Struct struct {
	Data map[*token.FileSet]*scanner.Error
}
`,
			pkgPaths: []string{"go/token", "go/scanner"},
		},
		{
			given: "given map from named type to another map from named type to the third named type",
			sourceCode: `
package ast
import (
"go/scanner"
"go/token"
"go/types"
)
type Struct struct {
	Data map[*token.FileSet]map[*scanner.Error]*types.Type
}
`,
			pkgPaths: []string{"go/token", "go/scanner", "go/types"},
		},
		{
			given: "given nested struct",
			sourceCode: `
package ast

type Struct struct {
	Field struct{
		Name string
	}
}
`,
			pkgPaths: []string{testPkgPath},
		},
		{
			given: "given nested interface",
			sourceCode: `
package ast

type Struct struct {
	Field interface {
		Name() string
	}
}
`,
			pkgPaths: []string{testPkgPath},
		},
		{
			given: "given builtin type",
			sourceCode: `
package ast

type Struct struct {
	Name string
}
`,
			pkgPaths: []string{},
		},
		{
			given: "given channel of named type",
			sourceCode: `
package ast
import "go/token" 
type Struct struct {
	FileSetChan chan *token.FileSet
}
`,
			pkgPaths: []string{"go/token"},
		},
		{
			given: "given func type",
			sourceCode: `
package ast
import (
"go/scanner"
"go/token"
"go/types"
)
type Struct struct {
	Func func(fileSet *token.FileSet, e *scanner.Error) (types.Type, error)
}
`,
			pkgPaths: []string{
				"go/token",
				"go/scanner",
				"go/types",
			},
		},
	}
	for _, row := range table {
		suite.testStructFieldPkgPath(row.given, row.sourceCode, row.pkgPaths...)
	}
}

func (suite *FieldInfoInternalTestSuite) testStructFieldPkgPath(given string, sourceCode string, pkgPaths ...string) {
	Convey(given, suite.T(), func() {
		suite.SetupTest()
		code := sourceCode
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, code)
			Convey("the field pkg path should be right", func() {
				So(err, ShouldBeNil)
				SoFieldPkgPathShouldEqual(suite.walker, pkgPaths...)
			})
		})
	})
}

func (suite *FieldInfoInternalTestSuite) TestFuncEmbeddedTypePkgPath() {
	Convey("given struct embed a named type struct", suite.T(), func() {
		sourceCode := `
package ast
import (
"go/types"
)
type Struct struct {
	types.Named
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("the typeInfo's dep pkg paths should be right", func() {
				So(err, ShouldBeNil)
				structInfo := suite.walker.Types()[0]
				depPkgPaths := structInfo.GetDepPkgPaths("")
				So(len(depPkgPaths), ShouldEqual, 1)
				So(depPkgPaths[0], ShouldEqual, "go/types")
			})
		})
	})
}

func SoFieldPkgPathShouldEqual(walker *typeWalker, pkgPaths ...string) {
	structInfo := walker.Types()[0]
	field1 := structInfo.Fields[0]
	packagePaths := field1.GetDepPkgPaths()
	So(lengthOfDepPkgPaths(field1), ShouldEqual, len(pkgPaths))
	for i, path := range packagePaths {
		So(path, ShouldEqual, pkgPaths[i])
	}
}
