package ast

type MethodInfo interface {
	GetName() string
	GetReceiver() TypeInfo
	GetParameterFields() []FieldInfo
	GetReturnFields() []FieldInfo
}

type methodInfo struct {
	receiver        TypeInfo
	parameterFields []FieldInfo
	returnFields    []FieldInfo
	name            string
}

func (m *methodInfo) GetParameterFields() []FieldInfo {
	return m.parameterFields
}

func (m *methodInfo) GetReturnFields() []FieldInfo {
	return m.returnFields
}

func (m *methodInfo) GetName() string {
	return m.name
}

func (m *methodInfo) GetReceiver() TypeInfo {
	return m.receiver
}
