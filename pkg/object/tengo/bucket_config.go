package tengo

import (
	"fmt"

	"github.com/aldor007/mort/pkg/config"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

type BucketConfig struct {
	tengoLib.ObjectImpl
	Value config.Bucket
}

func (o *BucketConfig) String() string {
	return o.Value.Name
}

func (o *BucketConfig) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

func (o *BucketConfig) IsFalsy() bool {
	return o.Value.Name == ""
}

func (o *BucketConfig) Equals(x tengoLib.Object) bool {
	other := x.(*BucketConfig)
	return o.Value.Name == other.Value.Name
}

func (o *BucketConfig) Copy() tengoLib.Object {

	return &BucketConfig{
		Value: o.Value,
	}
}

func (o *BucketConfig) TypeName() string {
	return "BucketConfig-object"
}

// IndexGet returns the value for the given key.
func (o *BucketConfig) IndexGet(index tengoLib.Object) (val tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	val = tengoLib.UndefinedValue
	switch strIdx {
	case "transform":
		if o.Value.Transform != nil {
			val = &Transform{Value: o.Value.Transform.ForParser()}
		} else {
			err = fmt.Errorf("no transform for %s", o.Value.Name)
			return
		}

	case "keys":
		keys := make([]tengoLib.Object, len(o.Value.Keys))
		for i, k := range o.Value.Keys {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["accessKey"] = &tengoLib.String{
				Value: k.AccessKey,
			}
			internalMap["secretAccessKey"] = &tengoLib.String{
				Value: k.SecretAccessKey,
			}
			keys[i] = &tengoLib.Map{Value: internalMap}
		}
		val = &tengoLib.Array{
			Value: keys,
		}
	case "headers":
		internalMap := make(map[string]tengoLib.Object)
		for k, v := range o.Value.Headers {
			internalMap[k] = &tengoLib.String{Value: v}
		}
		val = &tengoLib.Map{
			Value: internalMap,
		}
	case "name":
		val = &tengoLib.String{Value: o.Value.Name}

	}

	return
}
