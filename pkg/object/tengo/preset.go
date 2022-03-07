package tengo

import (
	"strings"

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
func (o *Preset) IndexGet(index tengoLib.Object) (res tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	var val tengoLib.Object
	switch strIdx {
	case "quality":
		val = &tengoLib.Int{Value: int64(o.Value.Quality)}
	case "format":
		val = &tengoLib.String{Value: o.Value.Format}
	}

	if strings.Contains(strIdx, "filters.") {
		parts := strings.Split(strIdx, ".")
		if len(parts) < 2 {
			return val, tengoLib.ErrInvalidIndexOnError
		}

		switch parts[1] {
		case "thumbnail":
			switch parts[2] {
			case "width":
				val = &tengoLib.Int{Value: int64(o.Value.Filters.Thumbnail.Width)}
			}
		}
	}

	return val, nil
}
