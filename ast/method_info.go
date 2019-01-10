package ast

type MethodInfo interface {
	GetName() string
	GetReceiver() TypeInfo
}

type methodInfo struct {
	receiver TypeInfo
	name     string
}

func (m *methodInfo) GetName() string {
	return m.name
}

func (m *methodInfo) GetReceiver() TypeInfo {
	return m.receiver
}
