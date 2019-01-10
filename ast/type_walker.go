package ast

import (
	"github.com/ahmetb/go-linq"
	"github.com/golang-collections/collections/stack"
	"github.com/lonegunmanb/johnnie"
	"go/ast"
	"go/types"
	"reflect"
)

type opsKind string

var analyzingType opsKind = "isAnalyzingType"
var analyzingFunc opsKind = "analyzingFunc"

type TypeWalker interface {
	johnnie.Walker
	GetTypes() []TypeInfo
	Parse(pkgPath string, sourceCode string) error
	ParseDir(dirPath string, ignorePattern string) error
	SetDir(dir string)
}

type typeWalker struct {
	abstractWalker
	types         []*typeInfo
	typeInfoStack stack.Stack
	opsStack      stack.Stack
}

func (walker *typeWalker) GetTypes() []TypeInfo {
	result := make([]TypeInfo, 0, len(walker.types))
	linq.From(walker.types).Select(func(t interface{}) interface{} {
		return t.(TypeInfo)
	}).ToSlice(&result)
	return result
}

func (walker *typeWalker) Types() []*typeInfo {
	return walker.types
}

func (walker *typeWalker) WalkFile(f *ast.File) {
	walker.pkgName = f.Name.Name
}

func (walker *typeWalker) WalkField(field *ast.Field) {
	if walker.isAnalyzingType() {
		typeInfo := walker.typeInfoStack.Peek().(*typeInfo)
		t := walker.analyzedTypes.Types[field.Type]
		fieldType := t.Type
		emitTypeNameIfFiledIsNestedType(walker, fieldType)
		typeInfo.processField(field, fieldType)
	}
}

func (walker *typeWalker) WalkStructType(structType *ast.StructType) {
	if walker.opsStack.Peek() == analyzingType {
		walker.addTypeInfo(structType, reflect.Struct)
	}
}

func (walker *typeWalker) EndWalkStructType(structType *ast.StructType) {
	walker.typeInfoStack.Pop()
}

func (walker *typeWalker) WalkInterfaceType(interfaceType *ast.InterfaceType) {
	if walker.opsStack.Peek() == analyzingType {
		walker.addTypeInfo(interfaceType, reflect.Interface)
	}
}

func (walker *typeWalker) EndWalkInterfaceType(interfaceType *ast.InterfaceType) {
	walker.typeInfoStack.Pop()
}

func (walker *typeWalker) WalkTypeSpec(spec *ast.TypeSpec) {
	walker.typeInfoStack.Push(spec.Name.Name)
	walker.opsStack.Push(analyzingType)
}

func (walker *typeWalker) EndWalkTypeSpec(spec *ast.TypeSpec) {
	walker.opsStack.Pop()
}

func (walker *typeWalker) WalkFuncType(funcType *ast.FuncType) {
	walker.opsStack.Push(analyzingFunc)
}

func (walker *typeWalker) EndWalkFuncType(funcType *ast.FuncType) {
	walker.opsStack.Pop()
}

func NewTypeWalker() TypeWalker {
	return newTypeWalkerWithPhysicalPath("")
}

func newTypeWalkerWithPhysicalPath(physicalPath string) TypeWalker {
	walker := &typeWalker{
		types: []*typeInfo{},
		abstractWalker: abstractWalker{
			osEnv:        NewGoPathEnv(),
			physicalPath: physicalPath,
		},
	}
	walker.abstractWalker.actualWalker = walker
	return walker
}

func (walker *typeWalker) addTypeInfo(typeExpr ast.Expr, kind reflect.Kind) {
	item := walker.typeInfoStack.Pop()
	typeName, ok := item.(string)
	if !ok {
		println(typeName)
	}
	analyzedType := walker.analyzedTypes.Types[typeExpr].Type
	typeInfo := &typeInfo{
		Name:         typeName,
		PkgPath:      walker.pkgPath,
		PkgName:      walker.pkgName,
		PhysicalPath: walker.physicalPath,
		Type:         analyzedType,
		Kind:         kind,
		declExp:      typeExpr,
	}
	walker.typeInfoStack.Push(typeInfo)
	walker.types = append(walker.types, typeInfo)
}

func (walker *typeWalker) isAnalyzingType() bool {
	return walker.opsStack.Peek() == analyzingType
}

func emitTypeNameIfFiledIsNestedType(walker *typeWalker, fieldType types.Type) {
	switch t := fieldType.(type) {
	case *types.Struct:
		{
			walker.typeInfoStack.Push(t.String())
		}
	case *types.Interface:
		{
			walker.typeInfoStack.Push(t.String())
		}
	case *types.Pointer:
		{
			emitTypeNameIfFiledIsNestedType(walker, t.Elem())
		}
	}
}

func isStructType(t types.Type) bool {
	_, ok := t.Underlying().(*types.Struct)
	return ok
}

func isEmbeddedField(field *ast.Field) bool {
	return field.Names == nil
}

func getTag(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	return field.Tag.Value
}
