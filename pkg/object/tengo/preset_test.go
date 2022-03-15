package tengo_test

import (
	"strings"
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

	assert.True(t, strings.Contains(tengoObject.String(), "format"))
	assert.False(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.True(t, !tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "Preset-object")
}

func TestPresetTengoGet(t *testing.T) {
	c := config.Preset{
		Quality: 100,
		Format:  "png",
		Filters: config.Filters{},
	}

	tengoObject := tengo.Preset{Value: c}

	// get quality
	v, err := tengoObject.IndexGet(&tengoLib.String{Value: "quality"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "int")
	qInt, _ := tengoLib.ToInt(v)
	assert.Equal(t, qInt, 100)

	// get format
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "format"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	fString, _ := tengoLib.ToString(v)
	assert.Equal(t, fString, "png")

	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "filters"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "Filters-object")
}
