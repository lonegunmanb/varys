package ast

import "go/types"

type MethodInfo interface {
	GetName() string
	GetReceiver() TypeInfo
	GetReturnTypes() []types.Type
}

type methodInfo struct {
	receiver    TypeInfo
	returnTypes []types.Type
	name        string
}

func (m *methodInfo) GetReturnTypes() []types.Type {
	return m.returnTypes
}

func (m *methodInfo) GetName() string {
	return m.name
}

func (m *methodInfo) GetReceiver() TypeInfo {
	return m.receiver
}
