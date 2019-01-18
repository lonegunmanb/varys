package ast

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go/types"
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
	testInputs := []*typeInfoTestData{
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
	for _, data := range testInputs {
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
	Given(testData.given, suite.T(), func() {
		suite.SetupTest()
		sourceCode := testData.sourceCode
		When("walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Then("struct name should equal to expected", func() {
				So(err, ShouldBeNil)
				typeInfo := suite.walker.Types()[0]
				And(typeInfo.Name, ShouldEqual, testData.structName)
				And(typeInfo.PkgPath, ShouldEqual, testPkgPath)
				And(typeInfo.PkgName, ShouldEqual, testData.pkgName)
			})
		})
	})
}

func (suite *typeInfoInternalTestSuite) TestGetStructDepImportPkgPaths() {
	Given("a complex struct", suite.T(), func() {
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
		When("walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Then("type's dep paths should equal to expected", func() {
				So(err, ShouldBeNil)
				structInfo := suite.walker.Types()[0]
				depPkgPaths := structInfo.GetDepPkgPaths("")
				And(depPkgPaths, ShouldHaveLength, 2)
				And(depPkgPaths, ShouldContain, "go/scanner")
				And(depPkgPaths, ShouldContain, "go/token")
			})
		})
	})
}

type stubFieldInfo struct {
	tag         string
	depPkgPaths []string
}

func (*stubFieldInfo) GetName() string {
	panic("implement me")
}

func (*stubFieldInfo) GetType() types.Type {
	panic("implement me")
}

func (f *stubFieldInfo) GetTag() string {
	return f.tag
}

func (*stubFieldInfo) GetReferenceFromType() TypeInfo {
	panic("implement me")
}

func (*stubFieldInfo) GetReferenceFromMethod() MethodInfo {
	panic("implement me")
}

func (f *stubFieldInfo) GetDepPkgPaths() []string {
	return f.depPkgPaths
}

func (suite *typeInfoInternalTestSuite) TestGetStructDepPkgPathsWithFieldTagFilter() {
	Given("a struct with two field which first one with a tag", suite.T(), func() {
		structInfo := &typeInfo{
			Fields: []FieldInfo{
				&stubFieldInfo{
					tag:         "inject:\"\"",
					depPkgPaths: []string{"go/scanner"},
				},
				&stubFieldInfo{
					tag:         "",
					depPkgPaths: []string{"go/types"},
				},
			},
		}
		When("get dep pkg paths with inject tag", func() {
			depPkgPaths := structInfo.GetDepPkgPaths("inject")
			Then("only go/scanner should be output", func() {
				And(depPkgPaths, ShouldHaveLength, 1)
				And(depPkgPaths[0], ShouldEqual, "go/scanner")
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
