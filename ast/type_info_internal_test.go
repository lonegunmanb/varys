package ast

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type typeInfoInternalTestSuite struct {
	suite.Suite
	walker *typeWalker
}

func TestTypeInfoTestSuite(t *testing.T) {
	suite.Run(t, &typeInfoInternalTestSuite{})
}

func (suite *typeInfoInternalTestSuite) SetupTest() {
	suite.walker = prepareTypeWalker(suite.T())
}

func (suite *typeInfoInternalTestSuite) TestStructInfo() {
	testDatas := []*typeInfoTestData{
		{
			given: "given simple struct code",
			sourceCode: `
	package ast
	type Struct struct{
	}
	`,
			structName: "Struct",
			pkgName:    "ast",
		},
		{
			given: "given struct with dep in same package",
			sourceCode: `
package ast
type Struct struct {
s Struct2
}
type Struct2 struct {

}
`,
			structName: "Struct",
			pkgName:    "ast",
		},
		{
			given: "given struct with package name different with pkg path",
			sourceCode: `
package test
type Struct struct{
}
`,
			structName: "Struct",
			pkgName:    "test",
		},
	}
	for _, data := range testDatas {
		suite.testStructInfo(data)
	}
}

type typeInfoTestData struct {
	given      string
	sourceCode string
	structName string
	pkgName    string
}

func (suite *typeInfoInternalTestSuite) testStructInfo(testData *typeInfoTestData) {
	Convey(testData.given, suite.T(), func() {
		suite.SetupTest()
		sourceCode := testData.sourceCode
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("struct name should equal to expected", func() {
				So(err, ShouldBeNil)
				typeInfo := suite.walker.Types()[0]
				So(typeInfo.Name, ShouldEqual, testData.structName)
				So(typeInfo.PkgPath, ShouldEqual, testPkgPath)
				So(typeInfo.PkgName, ShouldEqual, testData.pkgName)
			})
		})
	})
}

func (suite *typeInfoInternalTestSuite) TestGetStructDepImportPkgPaths() {
	Convey("given a complex struct", suite.T(), func() {
		sourceCode := `
package ast
import (
"go/scanner"
"go/token"
)
type Struct struct {
	Err scanner.Error //dep go/scanner
	FileSet token.FileSet //dep go/token
	PErr *scanner.Error //dep go/scanner
	PFileSet *token.FileSet //dep go/token
	SErr []scanner.Error //dep go/scanner
	SFileSet []token.FileSet //dep go/token
	AErr [1]scanner.Error //dep go/scanner
	AFileSet [1]token.FileSet //dep go/token
	NestedStruct struct { //dep pkgPath here
		Name string
	}
	NestedInterface interface { //dep pkgPath here
		Name() string
	}
	Map map[*scanner.Error]*token.FileSet //dep go/scanner, go/token
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("type's dep paths should equal to expected", func() {
				So(err, ShouldBeNil)
				structInfo := suite.walker.Types()[0]
				depPkgPaths := structInfo.GetDepPkgPaths("")
				So(len(depPkgPaths), ShouldEqual, 2)
				So(depPkgPaths, ShouldContain, "go/scanner")
				So(depPkgPaths, ShouldContain, "go/token")
			})
		})
	})
}

func TestIsNotTestFile(t *testing.T) {
	assert.True(t, isTestFile("rover_test.go"))
	assert.True(t, isTestFile("test.go"))
	assert.False(t, isTestFile("rover.go"))
}

func TestIsGoFile(t *testing.T) {
	assert.True(t, isGoSrcFile("src.go"))
	assert.False(t, isGoSrcFile("src.cpp"))
	assert.False(t, isGoSrcFile("go"))
}
