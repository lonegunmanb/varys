package ast

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterType(t *testing.T) {
	typeInterface := (*GoPathEnv)(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	defer ClearTypeRegister()
	mockOsEnv := NewMockGoPathEnv(ctrl)
	GetOrRegister(typeInterface, func() interface{} {
		return NewGoPathEnv()
	})
	RegisterType(typeInterface, func() interface{} {
		return mockOsEnv
	})
	actual := GetOrRegister(typeInterface, func() interface{} {
		return NewGoPathEnv()
	})
	assert.Equal(t, mockOsEnv, actual)
}
