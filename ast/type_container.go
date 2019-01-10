package ast

import (
	"fmt"
	"reflect"
)

var factories = make(map[string]func() interface{})

func GetOrRegister(interfaceType interface{}, factory func() interface{}) interface{} {
	typeName := getTypeName(interfaceType)
	f, ok := factories[typeName]
	if !ok {
		factories[typeName] = factory
		f = factory
	}
	return f()
}

func ClearTypeRegister() {
	factories = make(map[string]func() interface{})
}

func getTypeName(interfaceType interface{}) string {
	t := reflect.TypeOf(interfaceType).Elem()
	pkgPath := t.PkgPath()
	reflectedTypeName := t.Name()
	if pkgPath != "" {
		reflectedTypeName = fmt.Sprintf("%s.%s", pkgPath, reflectedTypeName)
	}
	return reflectedTypeName
}
