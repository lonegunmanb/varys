package ast

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
	"testing"
)

type fieldInfoInternalTestSuite struct {
	suite.Suite
	walker *typeWalker
}

func TestFieldInfoInternalTestSuite(t *testing.T) {
	suite.Run(t, new(fieldInfoInternalTestSuite))
}

func (suite *fieldInfoInternalTestSuite) SetupTest() {
	suite.walker = prepareTypeWalker(suite.T())
}

func (suite *fieldInfoInternalTestSuite) TestGetNamedTypeStructFieldPkgPath() {
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

func (suite *fieldInfoInternalTestSuite) TestStructFieldPkgPath() {
	table := []struct {
		given      string
		sourceCode string
		pkgPaths   []interface{}
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
			pkgPaths: []interface{}{"go/scanner"}},
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
			pkgPaths: []interface{}{"go/types"},
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
			pkgPaths: []interface{}{"go/scanner"},
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
			pkgPaths: []interface{}{"go/scanner"},
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
			pkgPaths: []interface{}{"go/token", "go/scanner"},
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
			pkgPaths: []interface{}{"go/token", "go/scanner", "go/types"},
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
			pkgPaths: []interface{}{testPkgPath},
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
			pkgPaths: []interface{}{testPkgPath},
		},
		{
			given: "given builtin type",
			sourceCode: `
package ast

type Struct struct {
	Name string
}
`,
			pkgPaths: []interface{}{},
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
			pkgPaths: []interface{}{"go/token"},
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
			pkgPaths: []interface{}{
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

func (suite *fieldInfoInternalTestSuite) testStructFieldPkgPath(given string, sourceCode string, pkgPaths ...interface{}) {
	Convey(given, suite.T(), func() {
		suite.SetupTest()
		code := sourceCode
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, code)
			Convey("the field pkg path should be right", func() {
				So(err, ShouldBeNil)
				So(suite.walker, shouldFieldPkgPathEqual, pkgPaths...)
			})
		})
	})
}

func (suite *fieldInfoInternalTestSuite) TestFuncEmbeddedTypePkgPath() {
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

func shouldFieldPkgPathEqual(actual interface{}, expected ...interface{}) string {
	walker := actual.(*typeWalker)
	structInfo := walker.Types()[0]
	field1 := structInfo.Fields[0]
	packagePaths := field1.GetDepPkgPaths()
	So(lengthOfDepPkgPaths(field1), ShouldEqual, len(expected))
	for i, path := range packagePaths {
		So(path, ShouldEqual, expected[i])
	}
	return ""
}
