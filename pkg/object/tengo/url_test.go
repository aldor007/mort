package tengo_test

import (
	"testing"

	"github.com/aldor007/mort/pkg/object/tengo"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
)

func TestUrlTengo(t *testing.T) {
	objectPath := "/bucket/image.png"
	c := pathToURL(objectPath)

	tengoObject := tengo.URL{Value: c}

	assert.Equal(t, tengoObject.String(), objectPath)
	assert.True(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.False(t, tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "url-object")
}

func TestUrlGetTengo(t *testing.T) {
	objectPath := "https://mort.com/bucket/image.png?width=100"
	c := pathToURL(objectPath)

	tengoObject := tengo.URL{Value: c}
	// get unknown index
	v, err := tengoObject.IndexGet(&tengoLib.String{Value: "no-name"})
	assert.Nil(t, err)
	assert.Equal(t, v, tengoLib.UndefinedValue)

	// invalid index type
	v, err = tengoObject.IndexGet(tengoLib.UndefinedValue)
	assert.Equal(t, err, tengoLib.ErrInvalidIndexType)

	// get scheme
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "scheme"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	schemaStr, _ := tengoLib.ToString(v)
	assert.Equal(t, schemaStr, "https")

	// get host
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "host"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	hostStr, _ := tengoLib.ToString(v)
	assert.Equal(t, hostStr, "mort.com")

	// get path
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "path"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	pathStr, _ := tengoLib.ToString(v)
	assert.Equal(t, pathStr, "/bucket/image.png")

	// get rawquery
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "rawquery"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	qsStr, _ := tengoLib.ToString(v)
	assert.Equal(t, qsStr, "width=100")

	// get query
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "query"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "map")
	tengoMap := v.(*tengoLib.Map)
	assert.Equal(t, len(tengoMap.Value), 1)
}
