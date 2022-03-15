package tengo

import (
	// "encoding/json"

	"github.com/aldor007/mort/pkg/config"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"gopkg.in/yaml.v2"
)

// Preset struct create Preset struct inside of tengo VM
type Preset struct {
	tengoLib.ObjectImpl
	Value config.Preset
}

// String return empty string
func (o *Preset) String() string {
	buf, _ := yaml.Marshal(&o.Value)
	return string(buf)
}

// BinaryOp not implemented
func (o *Preset) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

// IsFalsy return always true
func (o *Preset) IsFalsy() bool {
	return false
}

// Equals returns false
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
	case "filters":
		val = &Filters{Value: o.Value.Filters}
	}

	return val, nil
}
