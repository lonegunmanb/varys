package ast

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go/types"
	"reflect"
	"regexp"
	"testing"
)

//go:generate mockgen -package=ast -destination=./mock_type_walker.go github.com/lonegunmanb/varys/ast TypeWalker
func TestFuncDecl(t *testing.T) {
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
	typeWalker := parseCodeWithTypeWalker(t, sourceCode)
	struct1 := typeWalker.Types()[0]
	assert.Equal(t, 0, len(struct1.Fields))
}

func TestStructWithInterface(t *testing.T) {
	sourceCode := `
package ast
type TestStruct struct {
}
type TestInterface interface {
	Hello()
}
`
	typeWalker := parseCodeWithTypeWalker(t, sourceCode)
	assert.Equal(t, 2, len(typeWalker.Types()))
	testStruct := typeWalker.Types()[0]
	assert.Equal(t, 0, len(testStruct.Fields))
	assert.Equal(t, reflect.Struct, testStruct.Kind)
	testInterface := typeWalker.Types()[1]
	assert.Equal(t, 1, len(testInterface.Fields))
	assert.Equal(t, reflect.Interface, testInterface.Kind)
}

type TestStruct struct {
	Field1 int
	Field2 TestInterface
}
type TestInterface interface {
	Hello()
}

func TestAstTypeShouldEqualToReflectedType(t *testing.T) {
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
	typeWalker := parseCodeWithTypeWalker(t, sourceCode)
	testStruct := typeWalker.Types()[0]
	testStructInstance := TestStruct{}
	for i := 0; i < 2; i++ {
		assertFieldTypeNameEqual(t, testStruct, testStructInstance, i)
	}
}

func assertFieldTypeNameEqual(t *testing.T, testStruct *typeInfo, testStructInstance TestStruct, fieldIndex int) {
	astFieldTypeName := testStruct.Fields[fieldIndex].GetType().String()
	reflectedFieldType := reflect.TypeOf(testStructInstance).Field(fieldIndex).Type
	pkgPath := reflectedFieldType.PkgPath()
	reflectedTypeName := reflectedFieldType.Name()
	if notEmpty(pkgPath) {
		reflectedTypeName = fmt.Sprintf("%s.%s", pkgPath, reflectedTypeName)
	}
	assert.Equal(t, reflectedTypeName, astFieldTypeName)
}

func notEmpty(pkgPath string) bool {
	return pkgPath != ""
}

func TestFieldTag(t *testing.T) {
	souceCode := `
package ast
type Struct struct {
	Field2 int ` + "`" + "inject:\"Field2\"`" + `
}
`
	typeWalker := parseCodeWithTypeWalker(t, souceCode)
	struct1 := typeWalker.Types()[0]
	field := struct1.Fields[0]
	tag := field.GetTag()
	assert.Equal(t, "`inject:\"Field2\"`", tag)
}

func TestNewTypeDefinition(t *testing.T) {
	sourceCode := `
package ast
type newint int
type Struct struct {
	Field newint
	Field2 int
}
`
	typeWalker := parseCodeWithTypeWalker(t, sourceCode)
	struct1 := typeWalker.Types()[0]
	field1 := struct1.Fields[0]
	field2 := struct1.Fields[1]
	namedType, ok := field1.GetType().(*types.Named)
	assert.True(t, ok)
	assert.Equal(t, "newint", namedType.Obj().Name())
	assert.Equal(t, field2.GetType(), namedType.Underlying())
}

func TestTypeAliasIsIdenticalToType(t *testing.T) {
	sourceCode := `
package ast
type newint = int
type Struct struct {
	Field1 newint
	Field2 int
}
`
	typeWalker := parseCodeWithTypeWalker(t, sourceCode)
	struct1 := typeWalker.Types()[0]
	field1 := struct1.Fields[0]
	field2 := struct1.Fields[1]
	assert.Equal(t, field1.GetType(), field2.GetType())
}

func TestTypeFromImportWithDot(t *testing.T) {
	sourceCode := `
package ast
import . "go/ast"
type Struct1 struct {
	Field *Decl
}
`
	typeWalker := parseCodeWithTypeWalker(t, sourceCode)
	structs := typeWalker.Types()
	struct1 := structs[0]
	field := struct1.Fields[0]
	ft := field.GetType()
	d := ft.String()
	d2 := ft.Underlying().String()
	println(d)
	println(d2)
	assert.Equal(t, "*go/ast.Decl", field.GetType().String())
	//assert.Equal(t, "ast", fieldInfo.GetPkgName())
}

func TestWalkFieldInfos(t *testing.T) {
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
	typeWalker := parseCodeWithTypeWalker(t, sourceCode)
	structs := typeWalker.Types()
	struct1 := structs[0]
	assert.Equal(t, 6, len(struct1.Fields))
	field1 := struct1.Fields[0]
	assert.Equal(t, "Field1", field1.GetName())
	assert.Equal(t, "int", field1.GetType().String())
	field2 := struct1.Fields[1]
	assert.Equal(t, "Field2", field2.GetName())
	namedType, ok := field2.GetType().(*types.Named)
	assert.True(t, ok)
	assert.Equal(t, "Struct2", namedType.Obj().Name())
	struct2Type, ok := namedType.Underlying().(*types.Struct)
	assert.True(t, ok)
	assert.Equal(t, structs[1].Type, struct2Type)
	field3 := struct1.Fields[2]
	pointer, ok := field3.GetType().(*types.Pointer)
	assert.True(t, ok)
	assert.Equal(t, "int", pointer.Elem().String())
	field4 := struct1.Fields[3]
	field5 := struct1.Fields[4]
	assert.Equal(t, field4.GetType(), field5.GetType())
	float64Type, ok := field4.GetType().(*types.Basic)
	assert.True(t, ok)
	assert.Equal(t, "float64", float64Type.String())
	field6 := struct1.Fields[5]
	assert.Equal(t, "*go/ast.Decl", field6.GetType().String())
}

func TestWalkStructNames(t *testing.T) {
	const structName1 = "Struct"
	const structName2 = "Struct2"
	const field1Name = "MaleOne"
	const field2Name = "Field2"
	var structDefine1 = fmt.Sprintf(`
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

	typeWalker := parseCodeWithTypeWalker(t, structDefine1)
	structs := typeWalker.Types()
	assert.Len(t, structs, 2)
	structInfo := structs[0]
	assert.Equal(t, structName1, structInfo.Name)
	structInfo = structs[1]
	assert.Equal(t, structName2, structInfo.Name)
}

func TestWalkStructWithNestedStruct(t *testing.T) {
	const structDefine = `
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

	typeWalker := parseCodeWithTypeWalker(t, structDefine)
	walkedTypes := typeWalker.Types()
	assert.Len(t, walkedTypes, 4)
	structInfo := walkedTypes[0]
	assert.Equal(t, "NestedStruct", structInfo.Name)
	structInfo = walkedTypes[1]
	assert.Equal(t, "struct{MiddleStructField struct{Field1 int; Field2 string; BottomStructField struct{Field1 int; Field2 string}}}", structInfo.Name)
	structInfo = walkedTypes[2]
	assert.Equal(t, "struct{Field1 int; Field2 string; BottomStructField struct{Field1 int; Field2 string}}", structInfo.Name)
	structInfo = walkedTypes[3]
	assert.Equal(t, "struct{Field1 int; Field2 string}", structInfo.Name)
}

func TestWalkStructWithNestedInterface(t *testing.T) {
	const structDefine = `
package test
type Struct struct {
	Field interface {
		Name() string
	}
}
`
	typeWalker := parseCodeWithTypeWalker(t, structDefine)
	walkedTypes := typeWalker.Types()
	assert.Equal(t, 2, len(walkedTypes))
	nestedType := walkedTypes[1]
	assert.Equal(t, "interface{Name() string}", nestedType.Name)
}

func TestWalkStructWithNestedStructByPointer(t *testing.T) {
	testNestedStructWithStar(t, "*")
	testNestedStructWithStar(t, "**")
	testNestedStructWithStar(t, "***")
}

func testNestedStructWithStar(t *testing.T, star string) {
	const structDefine = `
	package test
	type Struct struct {
		Field %sstruct{Name string}
	}
	`
	typeWalker := parseCodeWithTypeWalker(t, fmt.Sprintf(structDefine, star))
	walkedTypes := typeWalker.Types()
	assert.Equal(t, 2, len(walkedTypes))
	nestedType := walkedTypes[1]
	assert.Equal(t, "struct{Name string}", nestedType.Name)
}

func TestWalkStructInheritingAnotherStruct(t *testing.T) {
	const structDefine = `
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
	typeWalker := parseCodeWithTypeWalker(t, structDefine)
	interface1 := typeWalker.Types()[0]
	struct1 := typeWalker.Types()[1]
	struct2 := typeWalker.Types()[2]
	struct3 := typeWalker.Types()[3]
	assert.Equal(t, 3, len(struct3.EmbeddedTypes))
	assertType(t, interface1, interface1.GetFullName(), EmbeddedByInterface, struct3.EmbeddedTypes[0])
	assertType(t, struct1, struct1.GetFullName(), EmbeddedByStruct, struct3.EmbeddedTypes[1])
	assertType(t, struct2, "*"+struct2.GetFullName(), EmbeddedByPointer, struct3.EmbeddedTypes[2])
}

func TestWalkerWithPhysicalPath(t *testing.T) {
	sourceCode := `
package ast
type TestStruct struct {
}
`
	expectedPhysicalPath := "expected"
	typeWalker := NewTypeWalker().(*typeWalker)
	typeWalker.physicalPath = expectedPhysicalPath
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOsEnv := NewMockGoPathEnv(ctrl)
	mockOsEnv.EXPECT().GetPkgPath(gomock.Eq(expectedPhysicalPath)).Times(1).Return(testPkgPath, nil)
	typeWalker.osEnv = mockOsEnv
	err := typeWalker.Parse(testPkgPath, sourceCode)
	assert.Nil(t, err)
	assert.Equal(t, expectedPhysicalPath, typeWalker.GetTypes()[0].GetPhysicalPath())
}

func assertType(t *testing.T, typeInfo *typeInfo, fullName string, expectedKind EmbeddedKind, embeddedType EmbeddedType) {
	assert.Equal(t, fullName, embeddedType.GetFullName())
	assert.Equal(t, typeInfo.PkgPath, embeddedType.GetPkgPath())
	assert.Equal(t, expectedKind, embeddedType.GetKind())
	assert.Equal(t, "", embeddedType.GetTag())
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
	assert.Panics(t, func() {
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
		returnField1(s1)
	})
}

func TestSubStructsWithSameStructureAreIdentical(t *testing.T) {
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
	s2 := Struct2{}
	s2.Field = s1.Field
	assert.Equal(t, 1, s2.Field.Field3.Field1)
	assert.Equal(t, "1", s2.Field.Field3.Field2)
}

func TestIgnorePattern(t *testing.T) {
	regex, err := regexp.Compile("mock_.*\\.go")
	assert.Nil(t, err)
	assert.True(t, regex.MatchString("mock_abc.go"))
}

func returnField1(input interface{}) string {
	return input.(Struct2).Field.Field1
}
