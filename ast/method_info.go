package ast

type MethodInfo interface {
	GetName() string
}

type methodInfo struct {
	name string
}

func (m *methodInfo) GetName() string {
	return m.name
}
