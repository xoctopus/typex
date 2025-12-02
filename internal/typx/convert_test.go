package typx_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/typx"
)

func TestNewTTByRT(t *testing.T) {
	for _, c := range LitTypeCases {
		fromRT := typx.NewLitType(typx.NewTTByRT(c.rt)).String()
		fromTT := typx.NewLitType(c.tt).String()
		Expect(t, fromRT, Equal(fromTT))
	}
}
