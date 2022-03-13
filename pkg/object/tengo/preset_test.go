package tengo_test

import (
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object/tengo"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
)

func TestPresetTengo(t *testing.T) {
	c := config.Preset{}

	tengoObject := tengo.Preset{Value: c}

	assert.Equal(t, tengoObject.String(), "")
	assert.False(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.True(t, tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "Preset-object")
}
