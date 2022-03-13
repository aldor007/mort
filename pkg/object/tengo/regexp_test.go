package tengo_test

import (
	"regexp"
	"testing"

	"github.com/aldor007/mort/pkg/object/tengo"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
)

func TestRegexpTengo(t *testing.T) {
	c := regexp.MustCompile("[a-z]+")

	tengoObject := tengo.Regexp{Value: c}

	assert.Equal(t, tengoObject.String(), "[a-z]+")
	assert.True(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.False(t, tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "Regexp-object")
}

func TestRegexpTengoCall(t *testing.T) {
	c := regexp.MustCompile("(?P<presetName>[a-z]+)")

	tengoObject := tengo.Regexp{Value: c}

	urlTengo := tengoLib.String{Value: "somethingAAA"}
	_, err := tengoObject.Call()
	assert.Equal(t, err, tengoLib.ErrWrongNumArguments)

	r, err := tengoObject.Call(&urlTengo)
	assert.Nil(t, err)
	assert.Equal(t, r.TypeName(), "immutable-map")

	rMap := r.(*tengoLib.ImmutableMap)

	presetName, err := rMap.IndexGet(&tengoLib.String{Value: "presetName"})
	assert.Nil(t, err)

	presetNameStr, _ := tengoLib.ToString(presetName)
	assert.Equal(t, presetNameStr, "something")

}
