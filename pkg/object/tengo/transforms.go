package tengo

import (

	"github.com/aldor007/mort/pkg/transforms"
	"github.com/d5/tengo/v2"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)


type Transforms struct {
    tengoLib.ObjectImpl
    Value transforms.Transforms
}

func (o *Transforms) String() string {
    return string(o.Value.Hash().Sum64())
}

func (o *Transforms) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
    return nil, tengoLib.ErrInvalidOperator
}

func (o *Transforms) IsFalsy() bool {
    return o.Value.NotEmpty
}

func (o *Transforms) Equals(x tengoLib.Object) bool {
	return false
}

func (o *Transforms) Copy() tengoLib.Object {
    return &Transforms{
        Value: o.Value,
    }
}

func (o *Transforms) TypeName() string {
    return "Transforms-object"
}
// IndexGet returns the value for the given key.
func (o *Transforms) IndexGet(index tengoLib.Object) (val tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	val = tengo.UndefinedValue
    switch strIdx {
    case "resize":
        val = &tengoLib.UserFunction{Name: "resize", Value: o.resize}
    case "extract":
        val = &tengoLib.UserFunction{Name: "extract", Value: o.extract}
    case "crop":
        val = &tengoLib.UserFunction{Name: "crop", Value: o.crop}
    case "resizeCropAuto":
        val = &tengoLib.UserFunction{Name: "resizeCropAuto", Value: o.resizeCropAuto}
    case "interlace":
        val = &tengoLib.UserFunction{Name: "interlace", Value: o.interlace}
    case "quality":
        val = &tengoLib.UserFunction{Name: "quality", Value: o.quality}
    case "stripMetadata":
        val = &tengoLib.UserFunction{Name: "stripMetadata", Value: o.stripMetadata}
    case "blur":
        val = &tengoLib.UserFunction{Name: "blur", Value: o.blur}
    case "format":
        val = &tengoLib.UserFunction{Name: "format", Value: o.format}
    case "watermark":
        val = &tengoLib.UserFunction{Name: "watermark", Value: o.watermark}
    case "grayscale":
        val = &tengoLib.UserFunction{Name: strIdx, Value: o.grayscale}
    case "rotate":
        val = &tengoLib.UserFunction{Name: strIdx, Value: o.rotate}
    }


	return val, nil
}

func (o *Transforms) resize(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 5 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var width, height int
	var enlarge, preserveAspectRatio, fill, ok bool
	if width, ok = tengoLib.ToInt(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: args[0].TypeName()}
	}

	if height, ok = tengo.ToInt(args[1]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: args[1].TypeName()}
	}

	if enlarge, ok = tengoLib.ToBool(args[2]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: args[2].TypeName()}
	}

	if preserveAspectRatio, ok = tengo.ToBool(args[3]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: args[3].TypeName()}
	}

	if fill, ok = tengo.ToBool(args[4]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: args[4].TypeName()}
	}


	o.Value.Resize(width, height, enlarge, preserveAspectRatio, fill)
	return tengo.UndefinedValue, nil
}

func (o *Transforms) extract(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 4 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var ok bool
	var top, left, width, height int
	if top, ok = tengoLib.ToInt(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: args[0].TypeName()}
	}

	if left, ok = tengo.ToInt(args[1]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: args[1].TypeName()}
	}
	if width, ok = tengoLib.ToInt(args[2]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: args[2].TypeName()}
	}

	if height, ok = tengo.ToInt(args[3]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: args[3].TypeName()}
	}



	o.Value.Extract(top, left, width, height)
	return tengo.UndefinedValue, nil
}

func (o *Transforms) crop(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 5 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var gravity string
	var ok, enlarge, embed  bool
	var width, height int
	if width, ok = tengoLib.ToInt(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: args[0].TypeName()}
	}

	if height, ok = tengo.ToInt(args[1]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: args[1].TypeName()}
	}

	if gravity, ok = tengo.ToString(args[2]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "gravity", Expected: "string", Found: args[2].TypeName()}
	}

	if enlarge, ok = tengo.ToBool(args[3]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "enlarge", Expected: "bool", Found: args[3].TypeName()}
	}

	if embed, ok = tengo.ToBool(args[4]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "enlarge", Expected: "bool", Found: args[4].TypeName()}
	}

	o.Value.Crop(width, height, gravity, enlarge, embed)
	return tengo.UndefinedValue, nil
}

func (o *Transforms) resizeCropAuto(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 2 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var ok bool
	var width, height int
	if width, ok = tengoLib.ToInt(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: args[0].TypeName()}
	}

	if height, ok = tengo.ToInt(args[1]); !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: args[1].TypeName()}
	}

	o.Value.ResizeCropAuto(width, height)
	return tengo.UndefinedValue, nil
}

func (o *Transforms) interlace(_...tengoLib.Object) (ret tengoLib.Object, err error) {

	o.Value.Interlace()
	return tengo.UndefinedValue, nil
}

func (o *Transforms) quality(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 1 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var ok bool
	var quality int
	if quality, ok = tengoLib.ToInt(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "quality", Expected: "int", Found: args[0].TypeName()}
	}

	o.Value.Quality(quality)
	return tengo.UndefinedValue, nil
}

func (o *Transforms) stripMetadata(_ ...tengoLib.Object) (ret tengoLib.Object, err error) {
	o.Value.StripMetadata()
	return tengo.UndefinedValue, nil
}

func (o *Transforms) blur(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 2 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var ok bool
	var sigma, minAmpl float64
	if sigma, ok = tengoLib.ToFloat64(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "sigma", Expected: "float64", Found: args[0].TypeName()}
	}

	if minAmpl, ok = tengoLib.ToFloat64(args[1]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "quality", Expected: "flaot64", Found: args[1].TypeName()}
	}

	o.Value.Blur(sigma, minAmpl)
	return tengo.UndefinedValue, nil
}

func (o *Transforms) format(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 1 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var ok bool
	var format string
	if format, ok = tengoLib.ToString(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "format", Expected: "string", Found: args[0].TypeName()}
	}

	return tengo.UndefinedValue, o.Value.Format(format)
}

func (o *Transforms) watermark(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 3 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var ok bool
	var image, position string
	var opacity float64
	if image, ok = tengoLib.ToString(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "image", Expected: "string", Found: args[0].TypeName()}
	}

	if position, ok = tengoLib.ToString(args[1]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "position", Expected: "string", Found: args[1].TypeName()}
	}

	if opacity, ok = tengoLib.ToFloat64(args[2]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "opacity", Expected: "float64", Found: args[2].TypeName()}
	}

	return tengo.UndefinedValue, o.Value.Watermark(image, position, float32(opacity))
}

func (o *Transforms) grayscale(_...tengoLib.Object) (ret tengoLib.Object, err error) {

	o.Value.Grayscale()
	return tengo.UndefinedValue, nil
}

func (o *Transforms) rotate(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 1 {
		return nil, tengoLib.ErrWrongNumArguments
	}

	var ok bool
	var angle int
	if angle, ok = tengoLib.ToInt(args[0]); !ok {
		return nil, tengoLib.ErrInvalidArgumentType{Name: "angle", Expected: "int", Found: args[0].TypeName()}
	}

	return tengo.UndefinedValue, o.Value.Rotate(angle)
}
