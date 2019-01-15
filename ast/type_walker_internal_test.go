package ast

import (
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go/types"
	"reflect"
	"regexp"
	"testing"
)

//go:generate mockgen -source=./type_walker.go -package=ast -destination=./mock_type_walker.go

type typeWalkerTestSuite struct {
	suite.Suite
	walker *typeWalker
}

func (suite *typeWalkerTestSuite) SetupTest() {
	suite.walker = prepareTypeWalker(suite.T())
}

func TestTypeWalkerSuite(t *testing.T) {
	suite.Run(t, &typeWalkerTestSuite{})
}

func (suite *typeWalkerTestSuite) TestFuncDecl() {
	Convey("given empty struct", suite.T(), func() {
		sourceCode := `
package ast
type Struct struct {
}
func Test(input int) int{
	i := input
	print(input)
	print(i)
	return 1
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("struct should has no field", func() {
				So(err, ShouldBeNil)
				struct1 := suite.walker.Types()[0]
				So(struct1.Fields, ShouldBeEmpty)
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestStructWithInterface() {
	Convey("given struct and interface", suite.T(), func() {
		sourceCode := `
package ast
type TestStruct struct {
}
type TestInterface interface {
	Hello()
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("walker should has struct type and interface type", func() {
				So(err, ShouldBeNil)
				walkedTypes := suite.walker.Types()
				So(walkedTypes, ShouldHaveLength, 2)
				structType := walkedTypes[0]
				interfaceType := walkedTypes[1]
				So(structType.Fields, ShouldBeEmpty)
				So(structType.Kind, ShouldEqual, reflect.Struct)
				So(len(interfaceType.Fields), ShouldEqual, 1)
				So(interfaceType.Kind, ShouldEqual, reflect.Interface)
			})
		})
	})
}

type TestStruct struct {
	Field1 int
	Field2 TestInterface
}
type TestInterface interface {
	Hello()
}

func (suite *typeWalkerTestSuite) TestAstTypeShouldEqualToReflectedType() {
	Convey("given struct type and interface type", suite.T(), func() {
		sourceCode := `
package ast
type TestStruct struct {
	Field1 int
	Field2 TestInterface
}
type TestInterface interface {
	Hello()
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("parsed type infos should equal to expected", func() {
				So(err, ShouldBeNil)
				testStructInfo := suite.walker.Types()[0]
				testStructInstance := TestStruct{}
				So(testStructInfo, shouldEqualToActualType, testStructInstance)
			})
		})
	})
}

func shouldEqualToActualType(actual interface{}, expected ...interface{}) string {
	testStruct := actual.(*typeInfo)
	actualStruct := expected[0].(TestStruct)
	for i := 0; i < 2; i++ {
		astFieldTypeName := testStruct.Fields[i].GetType().String()
		reflectedFieldType := reflect.TypeOf(actualStruct).Field(i).Type
		pkgPath := reflectedFieldType.PkgPath()
		reflectedTypeName := reflectedFieldType.Name()
		if notEmpty(pkgPath) {
			reflectedTypeName = fmt.Sprintf("%s.%s", pkgPath, reflectedTypeName)
		}
		So(astFieldTypeName, ShouldEqual, reflectedTypeName)
	}
	return ""
}

func notEmpty(pkgPath string) bool {
	return pkgPath != ""
}

func (suite *typeWalkerTestSuite) TestFieldTag() {
	Convey("given field with inject tag", suite.T(), func() {
		sourceCode := `
package ast
type Struct struct {
	Field2 int ` + "`" + "inject:\"Field2\"`" + `
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("field's tag info should equal to expected", func() {
				So(err, ShouldBeNil)
				struct1 := suite.walker.Types()[0]
				field := struct1.Fields[0]
				tag := field.GetTag()
				So(tag, ShouldEqual, "`inject:\"Field2\"`")
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestNewTypeDefinition() {
	Convey("given new type definition", suite.T(), func() {
		sourceCode := `
package ast
type newint int
type Struct struct {
	Field newint
	Field2 int
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("new type should has same type but different name", func() {
				So(err, ShouldBeNil)
				struct1 := suite.walker.Types()[0]
				field1 := struct1.Fields[0]
				field2 := struct1.Fields[1]
				namedType, ok := field1.GetType().(*types.Named)
				So(ok, ShouldBeTrue)
				So(namedType.Obj().Name(), ShouldEqual, "newint")
				So(namedType.Underlying(), ShouldEqual, field2.GetType())
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestTypeAliasIsIdenticalToType() {
	Convey("given type alias", suite.T(), func() {
		sourceCode := `
package ast
type newint = int
type Struct struct {
	Field1 newint
	Field2 int
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("two type are exactly same", func() {
				So(err, ShouldBeNil)
				structType := suite.walker.Types()[0]
				So(structType.Fields[0].GetType(), ShouldEqual, structType.Fields[1].GetType())
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestTypeFromImportWithDot() {
	Convey("given import with dot decl", suite.T(), func() {
		sourceCode := `
package ast
import . "go/ast"
type Struct1 struct {
	Field *Decl
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("field type info should equal to expected", func() {
				So(err, ShouldBeNil)
				structInfo := suite.walker.Types()[0]
				field := structInfo.Fields[0]
				So(field.GetType().String(), ShouldEqual, "*go/ast.Decl")
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestWalkFieldInfos() {
	Convey("given multiple fields struct", suite.T(), func() {
		sourceCode := `
package test
import "go/ast"
type Struct1 struct {
	Field1 int
	Field2 Struct2
	Field3 *int
	Field4, Field5 float64
	Field6 *ast.Decl
}

type Struct2 struct {
	
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("field infos should equal to expected", func() {
				So(err, ShouldBeNil)
				structInfo := suite.walker.GetTypes()[0]
				So(structInfo.GetFields(), ShouldHaveLength, 6)
				So(structInfo.GetFields()[0], shouldNameAndTypeEqual, "Field1", "int")
				So(structInfo.GetFields()[1], shouldNameAndTypeEqual, "Field2",
					"github.com/lonegunmanb/varys/ast.Struct2")
				So(structInfo.GetFields()[2], shouldNameAndTypeEqual, "Field3", "*int")
				So(structInfo.GetFields()[3], shouldNameAndTypeEqual, "Field4", "float64")
				So(structInfo.GetFields()[4], shouldNameAndTypeEqual, "Field5", "float64")
				So(structInfo.GetFields()[5], shouldNameAndTypeEqual, "Field6", "*go/ast.Decl")
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestWalkStructNames() {
	Convey("given two structs", suite.T(), func() {
		const structName1 = "Struct"
		const structName2 = "Struct2"
		const field1Name = "MaleOne"
		const field2Name = "Field2"
		sourceCode := fmt.Sprintf(`
package test
type %s struct {
	%s int
	%s string
}
type %s struct{
		%s string
		%s int
}
`, structName1,
			field1Name,
			field2Name,

			structName2,
			field1Name,
			field2Name)
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("struct names should equal to expected", func() {
				So(err, ShouldBeNil)
				structs := suite.walker.GetTypes()
				So(structs, ShouldHaveLength, 2)
				So(structs[0].GetName(), ShouldEqual, structName1)
				So(structs[1].GetName(), ShouldEqual, structName2)
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestWalkStructWithNestedStruct() {
	Convey("given source code with nested structs", suite.T(), func() {
		const sourceCode = `
package test
type NestedStruct struct {
	TopStructField struct {
		MiddleStructField struct {
			Field1 int
			Field2 string
			BottomStructField struct {
				Field1 int
				Field2 string
			}
		}
	}
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("all structs are properly parsed", func() {
				So(err, ShouldBeNil)
				walkedTypes := suite.walker.Types()
				So(walkedTypes, ShouldHaveLength, 4)
				So(walkedTypes[0].Name, ShouldEqual, "NestedStruct")
				So(walkedTypes[1].Name, ShouldEqual,
					"struct{MiddleStructField struct{Field1 int; Field2 string; BottomStructField struct{Field1 int; Field2 string}}}")
				So(walkedTypes[2].Name, ShouldEqual, "struct{Field1 int; Field2 string; BottomStructField struct{Field1 int; Field2 string}}")
				So(walkedTypes[3].Name, ShouldEqual, "struct{Field1 int; Field2 string}")
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestWalkStructWithNestedInterface() {
	Convey("given struct with nested interface", suite.T(), func() {
		const sourceCode = `
package test
type Struct struct {
	Field interface {
		Name() string
	}
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("interface type is properly parsed", func() {
				So(err, ShouldBeNil)
				walkerTypes := suite.walker.Types()
				So(walkerTypes, ShouldHaveLength, 2)
				So(walkerTypes[1].Name, ShouldEqual, "interface{Name() string}")
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestWalkStructWithNestedStructByPointer() {
	testNestedStructWithStar(suite, "*")
	testNestedStructWithStar(suite, "**")
	testNestedStructWithStar(suite, "***")
}

func testNestedStructWithStar(suite *typeWalkerTestSuite, star string) {
	Convey(fmt.Sprintf("given field type with start mark %s", star), suite.T(), func() {
		suite.SetupTest()
		const structDefine = `
	package test
	type Struct struct {
		Field %sstruct{Name string}
	}
	`
		sourceCode := fmt.Sprintf(structDefine, star)
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("nested struct type is properly parsed", func() {
				So(err, ShouldBeNil)
				walkedTypes := suite.walker.Types()
				So(walkedTypes, ShouldHaveLength, 2)
				So(walkedTypes[1].Name, ShouldEqual, "struct{Name string}")
			})
		})
	})
}

func (suite *typeWalkerTestSuite) TestWalkStructHasEmbeddingTypes() {
	Convey("given struct embedding two structs", suite.T(), func() {
		const sourceCode = `
package test
type Interface interface {
	Hello()
}
type Struct1 struct {
	FullName string
}
type Struct2 struct {
	Age int
}
type Struct3 struct {
	Interface
	Struct1
	*Struct2
}
`
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("embedded infos are properly parsed", func() {
				So(err, ShouldBeNil)
				typeWalker := suite.walker
				interface1 := typeWalker.Types()[0]
				struct1 := typeWalker.Types()[1]
				struct2 := typeWalker.Types()[2]
				struct3 := typeWalker.Types()[3]
				So(struct3.EmbeddedTypes, ShouldHaveLength, 3)
				So(struct3.EmbeddedTypes[0], shouldEqualEmbeddedType, EmbeddedByInterface, interface1.GetFullName(),
					testPkgPath)
				So(struct3.EmbeddedTypes[1], shouldEqualEmbeddedType, EmbeddedByStruct, struct1.GetFullName(),
					testPkgPath)
				So(struct3.EmbeddedTypes[2], shouldEqualEmbeddedType, EmbeddedByPointer, "*"+struct2.GetFullName(),
					testPkgPath)
			})
		})
	})
}

func shouldEqualEmbeddedType(actual interface{}, expected ...interface{}) string {
	embeddedType := actual.(EmbeddedType)
	expectedKind := expected[0].(EmbeddedKind)
	fullName := expected[1].(string)
	pkgPath := expected[2].(string)
	So(embeddedType.GetFullName(), ShouldEqual, fullName)
	So(embeddedType.GetPkgPath(), ShouldEqual, pkgPath)
	So(embeddedType.GetKind(), ShouldEqual, expectedKind)
	So(embeddedType.GetTag(), ShouldEqual, "")
	return ""
}

func (suite *typeWalkerTestSuite) TestWalkerWithPhysicalPath() {
	Convey("given gopath different with default", suite.T(), func() {
		sourceCode := `
package ast
type TestStruct struct {
}
`
		expectedPhysicalPath := "expected"
		suite.walker.physicalPath = expectedPhysicalPath
		suite.walker.osEnv.(*MockGoPathEnv).EXPECT().GetPkgPath(gomock.Eq(expectedPhysicalPath)).Times(1).Return(testPkgPath, nil)
		Convey("when walker parse source code", func() {
			err := suite.walker.Parse(testPkgPath, sourceCode)
			Convey("typeInfo's physical should equal to expected", func() {
				So(err, ShouldBeNil)
				So(suite.walker.GetTypes()[0].GetPhysicalPath(), ShouldEqual, expectedPhysicalPath)
			})
		})
	})
}

type Struct struct {
	Field struct {
		Field1 string
		Field2 int
		Field3 struct {
			Field1 int
			Field2 string
		}
	}
}

type Struct2 struct {
	Field struct {
		Field1 string
		Field2 int
		Field3 struct {
			Field1 int
			Field2 string
		}
	}
}

func TestDifferentNamedStructCannotConvert(t *testing.T) {
	Convey("given an anonymous struct with very same structure with Struct2", t, func() {
		s1 := Struct{
			Field: struct {
				Field1 string
				Field2 int
				Field3 struct {
					Field1 int
					Field2 string
				}
			}{
				Field1: "1",
				Field2: 1,
				Field3: struct {
					Field1 int
					Field2 string
				}{Field1: 1, Field2: "1"},
			},
		}
		Convey("when convert s1 to Struct2 type", func() {
			convert := func() {
				returnField1(s1)
			}
			Convey("should panic", func() {
				So(func() { convert() }, ShouldPanic)
			})
		})
	})
}

func TestSubStructsWithSameStructureAreIdentical(t *testing.T) {
	Convey("given a struct with nested struct same to struct2's nested struct", t, func() {
		s1 := Struct{
			Field: struct {
				Field1 string
				Field2 int
				Field3 struct {
					Field1 int
					Field2 string
				}
			}{
				Field1: "1",
				Field2: 1,
				Field3: struct {
					Field1 int
					Field2 string
				}{Field1: 1, Field2: "1"},
			},
		}
		Convey("when assign nested struct field to Struct2's nested struct field", func() {
			s2 := Struct2{}
			s2.Field = s1.Field
			Convey("two structs' nested struct fields are equal", func() {
				So(s2.Field, ShouldHaveSameTypeAs, s1.Field)
				So(s2.Field, shouldDeepEqual, s1.Field)
			})
		})
	})
}

func TestIgnorePattern(t *testing.T) {
	regex, err := regexp.Compile("mock_.*\\.go")
	assert.Nil(t, err)
	assert.True(t, regex.MatchString("mock_abc.go"))
}

func returnField1(input interface{}) string {
	return input.(Struct2).Field.Field1
}

func shouldNameAndTypeEqual(actual interface{}, expected ...interface{}) string {
	fieldInfo := actual.(FieldInfo)
	So(fieldInfo.GetName(), ShouldEqual, expected[0])
	So(fieldInfo.GetType().String(), ShouldEqual, expected[1])
	return ""
}

func shouldDeepEqual(actual interface{}, expected ...interface{}) string {
	if !reflect.DeepEqual(actual, expected[0]) {
		return "not equal"
	}
	return ""
}
