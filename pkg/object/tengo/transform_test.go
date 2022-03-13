package tengo_test

import (
	"regexp"
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object/tengo"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
)

func TestTransformTengo(t *testing.T) {

	c := &config.Transform{
		Path:          "/aaa",
		ParentStorage: "basic",
		ParentBucket:  "parent-bucket",
		PathRegexp:    regexp.MustCompile("[a-z]+"),
	}

	tengoObject := tengo.Transform{Value: c}

	assert.Equal(t, tengoObject.String(), "/aaa")
	assert.True(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.False(t, tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "Transform-object")
}

func TestTransformGetTengo(t *testing.T) {
	c := &config.Transform{
		Path:          "/aaa",
		ParentStorage: "basic",
		ParentBucket:  "parent-bucket",
		PathRegexp:    regexp.MustCompile("[a-z]+"),
		Kind:          "it",
	}

	tengoObject := tengo.Transform{Value: c}
	// get unknown index
	v, err := tengoObject.IndexGet(&tengoLib.String{Value: "no-name"})
	assert.Nil(t, err)
	assert.Equal(t, v, tengoLib.UndefinedValue)

	// invalid index type
	v, err = tengoObject.IndexGet(tengoLib.UndefinedValue)
	assert.Equal(t, err, tengoLib.ErrInvalidIndexType)

	// get path
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "path"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	pathStr, _ := tengoLib.ToString(v)
	assert.Equal(t, pathStr, "/aaa")

	// get parentStorage
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "parentStorage"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	pStorage, _ := tengoLib.ToString(v)
	assert.Equal(t, pStorage, "basic")

	// get parentBucket
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "parentBucket"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	pBucket, _ := tengoLib.ToString(v)
	assert.Equal(t, pBucket, "parent-bucket")

	// get regexp
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "pathRegexp"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "Regexp-object")

	// get kind
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "kind"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	kind, _ := tengoLib.ToString(v)
	assert.Equal(t, kind, "it")

	// get presets
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "presets"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "immutable-map")

}
