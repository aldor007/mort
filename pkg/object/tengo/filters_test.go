
package tengo_test

import (
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object/tengo"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
)

func TestFiltersTengo(t *testing.T) {
	c := config.Filters{}

	tengoObject := tengo.Filters{Value: c}

	assert.Equal(t, tengoObject.String(), "")
	assert.False(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.True(t, tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "Filters-object")
}

func TestFiltersTengoGetEmpty(t *testing.T) {

	c := config.Filters{}

	tengoObject := tengo.Filters{Value: c}

	valuesUndefine := []string{"thumbnail", "t","crop", "extract", "resizeCropAuto", "blur", "rotate", "watermark"}

	for _, v := range valuesUndefine {
		res, err := tengoObject.IndexGet(&tengoLib.String{Value: v})
		assert.Nil(t, err)
		assert.Equal(t, res, tengoLib.UndefinedValue)
	}
	valuesFalse:= []string{"strip", "grayscale", "autoRotate", "interlace"}

	for _, v := range valuesFalse {
		res, err := tengoObject.IndexGet(&tengoLib.String{Value: v})
		assert.Nil(t, err)
		assert.Equal(t, res, tengoLib.FalseValue)
	}

}

func TestFiltersTengoGet(t *testing.T) {

	c := config.Filters{
		Thumbnail:  &struct{Width int "yaml:\"width\""; Height int "yaml:\"height\""; Mode string "yaml:\"mode\""; PreserveAspectRatio bool "yaml:\"preserveAspectRatio\""; Fill bool "yaml:\"fill\""}{
			Width:  100,
			Height: 100,
		},
		Crop: &struct{Width int "yaml:\"width\""; Height int "yaml:\"height\""; Gravity string "yaml:\"gravity\""; Mode string "yaml:\"mode\""; Embed bool "yaml:\"embed\""}{
			Width: 111,
			Height: 111,
			Gravity: "aaa",
		},
		Extract: &struct{Width int "yaml:\"width\""; Height int "yaml:\"height\""; Top int "yaml:\"top\""; Left int "yaml:\"left\""}{
			Width: 222,
			Height: 222,
			Top: 111,
			Left: 111,
		},
		ResizeCropAuto: &struct{Width int "yaml:\"width\""; Height int "yaml:\"height\""}{
			Width: 333,
			Height: 333,
		},
		Blur: &struct{Sigma float64 "yaml:\"sigma\""; MinAmpl float64 "yaml:\"minAmpl\""}{
			Sigma: 1.1,
			MinAmpl: 1.2,
		},
		Watermark: &struct{Image string "yaml:\"image\""; Position string "yaml:\"position\""; Opacity float32 "yaml:\"opacity\""}{
			Image: "aaa.png",
			Position: "top-left",
			Opacity: 2.2,
		},
		Rotate: &struct{Angle int "yaml:\"angle\""}{
			Angle: 289,
		},

	}

	tengoObject := tengo.Filters{Value: c}

	res, err := tengoObject.IndexGet(&tengoLib.String{Value: "thumbnail"})
	assert.Nil(t, err)
	assert.Equal(t, res.TypeName(), "immutable-map")
	widthTengo, _ := res.IndexGet(&tengoLib.String{Value: "width"})
	width, _ := tengoLib.ToInt(widthTengo)
	assert.Equal(t, width, 100)

	res, err = tengoObject.IndexGet(&tengoLib.String{Value: "crop"})
	assert.Nil(t, err)
	assert.Equal(t, res.TypeName(), "immutable-map")
	widthTengo, _ = res.IndexGet(&tengoLib.String{Value: "width"})
	width, _ = tengoLib.ToInt(widthTengo)
	assert.Equal(t, width, 111)
	gravityTengo, _ := res.IndexGet(&tengoLib.String{Value: "gravity"})
	gravity, _ := tengoLib.ToString(gravityTengo)
	assert.Equal(t, gravity, "aaa")

	res, err = tengoObject.IndexGet(&tengoLib.String{Value: "extract"})
	assert.Nil(t, err)
	assert.Equal(t, res.TypeName(), "immutable-map")
	widthTengo, _ = res.IndexGet(&tengoLib.String{Value: "width"})
	width, _ = tengoLib.ToInt(widthTengo)
	assert.Equal(t, width, 222)

	res, err = tengoObject.IndexGet(&tengoLib.String{Value: "resizeCropAuto"})
	assert.Nil(t, err)
	assert.Equal(t, res.TypeName(), "immutable-map")
	widthTengo, _ = res.IndexGet(&tengoLib.String{Value: "width"})
	width, _ = tengoLib.ToInt(widthTengo)
	assert.Equal(t, width, 333)

	res, err = tengoObject.IndexGet(&tengoLib.String{Value: "resizeCropAuto"})
	assert.Nil(t, err)
	assert.Equal(t, res.TypeName(), "immutable-map")
	widthTengo, _ = res.IndexGet(&tengoLib.String{Value: "width"})
	width, _ = tengoLib.ToInt(widthTengo)
	assert.Equal(t, width, 333)

}