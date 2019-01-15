package ast

import (
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRegisterType(t *testing.T) {
	Convey("given container init with default instance", t, func() {
		typeInterface := (*GoPathEnv)(nil)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer ClearTypeRegister()
		expected := NewMockGoPathEnv(ctrl)
		GetOrRegister(typeInterface, func() interface{} {
			return NewGoPathEnv()
		})
		Convey("then register new factory with same key", func() {
			RegisterType(typeInterface, func() interface{} {
				return expected
			})
			Convey("when use GetOrRegister with origin default factory", func() {
				actual := GetOrRegister(typeInterface, func() interface{} {
					return NewGoPathEnv()
				})
				Convey("resolved instance should be equal to second registered instance", func() {
					So(actual, ShouldEqual, expected)
				})
			})
		})
	})
}
