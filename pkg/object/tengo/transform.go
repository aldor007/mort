package tengo

import (
	"github.com/aldor007/mort/pkg/config"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

type Transform struct {
	tengoLib.ObjectImpl
	Value *config.Transform
}

func (o *Transform) String() string {
	return o.Value.Path
}

func (o *Transform) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

func (o *Transform) IsFalsy() bool {
	return o.Value.Path == ""
}

func (o *Transform) Equals(x tengoLib.Object) bool {
	return false
}

func (o *Transform) Copy() tengoLib.Object {
	return &Transform{
		Value: o.Value,
	}
}

func (o *Transform) TypeName() string {
	return "Transform-object"
}

// IndexGet returns the value for the given key.
func (o *Transform) IndexGet(index tengoLib.Object) (res tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	var val tengoLib.Object
	switch strIdx {
	case "path":
		val = &tengoLib.String{Value: o.Value.Path}
	case "parentstorage":
		val = &tengoLib.String{Value: o.Value.ParentStorage}
	case "parentbucket":
		val = &tengoLib.String{Value: o.Value.ParentBucket}
	case "pathregexp":
		val = &Regexp{Value: o.Value.PathRegexp}
	case "kind":
		val = &tengoLib.String{Value: o.Value.Kind}
	case "presets":
		internalMap := make(map[string]tengoLib.Object)
		for k, v := range o.Value.Presets {
			internalMap[k] = &Preset{Value: v}
		}
		val = &tengoLib.Map{Value: internalMap}

	}

	return val, nil
}
