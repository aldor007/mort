package tengo

import (
	"github.com/aldor007/mort/pkg/config"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

type Preset struct {
	tengoLib.ObjectImpl
	Value config.Preset
}

func (o *Preset) String() string {
	return ""
}

func (o *Preset) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

func (o *Preset) IsFalsy() bool {
	return true
}

func (o *Preset) Equals(_ tengoLib.Object) bool {
	return false
}

func (o *Preset) Copy() tengoLib.Object {
	return &Preset{
		Value: o.Value,
	}
}

func (o *Preset) TypeName() string {
	return "Preset-object"
}

// IndexGet returns the value for the given key.
func (o *Preset) IndexGet(index tengoLib.Object) (val tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	val = tengoLib.UndefinedValue
	switch strIdx {
	case "quality":
		val = &tengoLib.Int{Value: int64(o.Value.Quality)}
	case "format":
		val = &tengoLib.String{Value: o.Value.Format}
	}

	return val, nil
}
