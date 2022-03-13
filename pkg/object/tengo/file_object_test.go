package tengo_test

import (
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/object/tengo"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
	"net/url"
)

func pathToURL(urlPath string) *url.URL {
	u, _ := url.Parse(urlPath)
	return u
}

func TestFileObjectTengo(t *testing.T) {
	objectPath := "/bucket/image.png"
	mortConfig := config.GetInstance()
	c, _ := object.NewFileObject(pathToURL(objectPath), mortConfig)

	tengoObject := tengo.FileObject{Value: c}

	assert.Equal(t, tengoObject.String(), objectPath)
	assert.True(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.False(t, tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "FileObject-object")
}

func TestFileObjectGetTengo(t *testing.T) {
	objectPath := "/bucket/image.png"
	mortConfig := config.GetInstance()
	c, _ := object.NewFileObject(pathToURL(objectPath), mortConfig)

	tengoObject := tengo.FileObject{Value: c}
	// get unknown index
	v, err := tengoObject.IndexGet(&tengoLib.String{Value: "no-name"})
	assert.Nil(t, err)
	assert.Equal(t, v, tengoLib.UndefinedValue)

	// invalid index type
	v, err = tengoObject.IndexGet(tengoLib.UndefinedValue)
	assert.Equal(t, err, tengoLib.ErrInvalidIndexType)

	// get uri
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "uri"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "url-object")
	assert.Equal(t, v.String(), objectPath)

	// get bucket
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "bucket"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	bStr, _ := tengoLib.ToString(v)
	assert.Equal(t, bStr, "bucket")

	// get transforms
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "transforms"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "Transforms-object")

	// get key
	v, err = tengoObject.IndexGet(&tengoLib.String{Value: "key"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	kStr, _ := tengoLib.ToString(v)
	assert.Equal(t, kStr, "/image.png")

}
