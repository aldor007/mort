package tengo

import (
	"github.com/aldor007/mort/pkg/config"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

// Filters struct create Filters struct inside of tengo VM
type Filters struct {
	tengoLib.ObjectImpl
	Value config.Filters
}

// String return empty string
func (o *Filters) String() string {
	return ""
}

// BinaryOp not implemented
func (o *Filters) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

// IsFalsy return always true
func (o *Filters) IsFalsy() bool {
	return true
}

// Equals returns false
func (o *Filters) Equals(_ tengoLib.Object) bool {
	return false
}

func (o *Filters) Copy() tengoLib.Object {
	return &Filters{
		Value: o.Value,
	}
}

func (o *Filters) TypeName() string {
	return "Filters-object"
}

// IndexGet returns the value for the given key.
func (o *Filters) IndexGet(index tengoLib.Object) (val tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	val = tengoLib.UndefinedValue
	switch strIdx {
	case "thumbnail":
		if o.Value.Thumbnail != nil {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["width"] = &tengoLib.Int{Value: int64(o.Value.Thumbnail.Width)}
			internalMap["height"] = &tengoLib.Int{Value: int64(o.Value.Thumbnail.Height)}
			internalMap["mode"] = &tengoLib.String{Value: o.Value.Thumbnail.Mode}
			if o.Value.Thumbnail.PreserveAspectRatio {
				internalMap["preserveAspectRatio"] = tengoLib.TrueValue
			} else {
				internalMap["preserveAspectRatio"] = tengoLib.FalseValue
			}
			if o.Value.Thumbnail.Fill {
				internalMap["fill"] = tengoLib.TrueValue
			} else {
				internalMap["fill"] = tengoLib.FalseValue
			}
			val = &tengoLib.ImmutableMap{Value: internalMap}
		}
	case "crop":
		if o.Value.Crop != nil {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["width"] = &tengoLib.Int{Value: int64(o.Value.Crop.Width)}
			internalMap["height"] = &tengoLib.Int{Value: int64(o.Value.Crop.Height)}
			internalMap["mode"] = &tengoLib.String{Value: o.Value.Crop.Mode}
			internalMap["gravity"] = &tengoLib.String{Value: o.Value.Crop.Gravity}
			val = &tengoLib.ImmutableMap{Value: internalMap}
			if o.Value.Crop.Embed {
				internalMap["embed"] = tengoLib.TrueValue
			} else {
				internalMap["embed"] = tengoLib.FalseValue
			}
		}
	case "extract":
		if o.Value.Extract != nil {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["width"] = &tengoLib.Int{Value: int64(o.Value.Extract.Width)}
			internalMap["height"] = &tengoLib.Int{Value: int64(o.Value.Extract.Height)}
			internalMap["top"] = &tengoLib.Int{Value: int64(o.Value.Extract.Top)}
			internalMap["left"] = &tengoLib.Int{Value: int64(o.Value.Extract.Left)}
			val = &tengoLib.ImmutableMap{Value: internalMap}
		}
	case "resizeCropAuto":
		if o.Value.ResizeCropAuto != nil {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["width"] = &tengoLib.Int{Value: int64(o.Value.ResizeCropAuto.Width)}
			internalMap["height"] = &tengoLib.Int{Value: int64(o.Value.ResizeCropAuto.Height)}
			val = &tengoLib.ImmutableMap{Value: internalMap}
		}
	case "blur":
		if o.Value.Blur != nil {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["sigma"] = &tengoLib.Float{Value: o.Value.Blur.Sigma}
			internalMap["minAmpl"] = &tengoLib.Float{Value: o.Value.Blur.MinAmpl}
			val = &tengoLib.ImmutableMap{Value: internalMap}
		}
	case "rotate":
		if o.Value.Rotate != nil {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["angle"] = &tengoLib.Int{Value: int64(o.Value.Rotate.Angle)}
			val = &tengoLib.ImmutableMap{Value: internalMap}
		}
	case "watermark":
		if o.Value.Watermark != nil {
			internalMap := make(map[string]tengoLib.Object)
			internalMap["image"] = &tengoLib.String{Value: o.Value.Watermark.Image}
			internalMap["position"] = &tengoLib.String{Value: o.Value.Watermark.Position}
			internalMap["opacity"] = &tengoLib.Float{Value: float64(o.Value.Watermark.Opacity)}
			val = &tengoLib.ImmutableMap{Value: internalMap}
		}
	case "interlace":
		if o.Value.Interlace {
			val = tengoLib.TrueValue
		} else {
			val = tengoLib.FalseValue
		}
	case "autoRotate":
		if o.Value.AutoRotate {
			val = tengoLib.TrueValue
		} else {
			val = tengoLib.FalseValue
		}
	case "grayscale":
		if o.Value.Grayscale {
			val = tengoLib.TrueValue
		} else {
			val = tengoLib.FalseValue
		}
	case "strip":
		if o.Value.Strip {
			val = tengoLib.TrueValue
		} else {
			val = tengoLib.FalseValue
		}
	}

	return val, nil
}
