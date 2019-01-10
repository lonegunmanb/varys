package ast

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVisitMethod(t *testing.T) {
	code := `
package ast
type Struct struct {
}

func (s *Struct) Method() {
}`
	walker := parseCodeWithFuncWalker(t, code)
	methods := walker.methods
	assert.Equal(t, 1, len(methods))
	assert.Equal(t, "Method", methods[0].GetName())
}
