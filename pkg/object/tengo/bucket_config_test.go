package tengo_test

import (
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object/tengo"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
)

func TestBucketConfigTengo(t *testing.T) {
	c := config.Bucket{
		Name:      "test",
		Transform: &config.Transform{},
		Headers:   nil,
	}
	tengoConfig := tengo.BucketConfig{Value: c}

	assert.Equal(t, tengoConfig.String(), "test")
	assert.True(t, tengoConfig.Equals(tengoConfig.Copy()))
	o, err := tengoConfig.BinaryOp(token.Add, tengoConfig.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.False(t, tengoConfig.IsFalsy())
	assert.Equal(t, tengoConfig.TypeName(), "BucketConfig-object")
}

func TestBucketConfigGetTengo(t *testing.T) {
	c := config.Bucket{
		Name:      "test",
		Transform: nil,
		Headers:   nil,
		Keys: []config.S3Key{
			config.S3Key{
				AccessKey:       "aaa",
				SecretAccessKey: "bbb",
			},
		},
	}
	tengoConfig := tengo.BucketConfig{Value: c}
	// get unknown index
	v, err := tengoConfig.IndexGet(&tengoLib.String{Value: "no-name"})
	assert.Nil(t, err)
	assert.Equal(t, v, tengoLib.UndefinedValue)

	// invalid index type
	v, err = tengoConfig.IndexGet(tengoLib.UndefinedValue)
	assert.Equal(t, err, tengoLib.ErrInvalidIndexType)

	// get transform - nil
	v, err = tengoConfig.IndexGet(&tengoLib.String{Value: "transform"})
	assert.NotNil(t, err)

	// get transform
	tengoConfig.Value.Transform = &config.Transform{}
	v, err = tengoConfig.IndexGet(&tengoLib.String{Value: "transform"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "Transform-object")

	// get keys
	v, err = tengoConfig.IndexGet(&tengoLib.String{Value: "keys"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "array")
	vArr := v.(*tengoLib.Array)
	assert.Equal(t, len(vArr.Value), len(c.Keys))
	k, err := vArr.IndexGet(&tengoLib.Int{Value: 0})
	assert.Nil(t, err)
	kSecret, err := k.IndexGet(&tengoLib.String{Value: "accessKey"})
	assert.Nil(t, err)
	kString, _ := tengoLib.ToString(kSecret)
	assert.Equal(t, kString, "aaa")

	// get headers - nil
	v, err = tengoConfig.IndexGet(&tengoLib.String{Value: "headers"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "map")

	// get headers
	tengoConfig.Value.Headers = make(map[string]string)
	tengoConfig.Value.Headers["x-test"] = "test"

	v, err = tengoConfig.IndexGet(&tengoLib.String{Value: "headers"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "map")
	hMap, err := v.IndexGet(&tengoLib.String{Value: "x-test"})
	assert.Nil(t, err)
	hValue, _ := tengoLib.ToString(hMap)
	assert.Equal(t, hValue, "test")

	// get name
	v, err = tengoConfig.IndexGet(&tengoLib.String{Value: "name"})
	assert.Nil(t, err)
	assert.Equal(t, v.TypeName(), "string")
	name, _ := tengoLib.ToString(v)
	assert.Equal(t, name, "test")

}
