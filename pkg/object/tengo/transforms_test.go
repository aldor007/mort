package tengo_test

import (
	"fmt"
	"testing"

	"github.com/aldor007/mort/pkg/object/tengo"
	"github.com/aldor007/mort/pkg/transforms"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"github.com/stretchr/testify/assert"
)

func TestTransformsTengo(t *testing.T) {
	c := transforms.Transforms{}

	tengoObject := tengo.Transforms{Value: c}

	assert.Equal(t, tengoObject.String(), "a692a0f768855173")
	assert.True(t, tengoObject.Equals(tengoObject.Copy()))
	o, err := tengoObject.BinaryOp(token.Add, tengoObject.Copy())
	assert.Nil(t, o)
	assert.Equal(t, err, tengoLib.ErrInvalidOperator)
	assert.False(t, tengoObject.IsFalsy())
	assert.Equal(t, tengoObject.TypeName(), "Transforms-object")
}

func TestTransformsIndexGet(t *testing.T) {
	methods := []string{
		"resize",
		"extract",
		"crop",
		"resizeCropAuto",
		"interlace",
		"quality",
		"stripMetadata",
		"blur",
		"format",
		"watermark",
		"grayscale",
		"rotate",
	}

	t.Run("methods", func(t *testing.T) {
		c := transforms.Transforms{}
		tengoObject := tengo.Transforms{Value: c}
		// get unknown index
		v, err := tengoObject.IndexGet(&tengoLib.String{Value: "no-name"})
		assert.Nil(t, err)
		assert.Equal(t, v, tengoLib.UndefinedValue)

		// invalid index type
		v, err = tengoObject.IndexGet(tengoLib.UndefinedValue)
		assert.Equal(t, err, tengoLib.ErrInvalidIndexType)
		for _, method := range methods {
			v, err := tengoObject.IndexGet(&tengoLib.String{Value: method})
			assert.Nil(t, err)
			assert.Equal(t, v.TypeName(), fmt.Sprintf("user-function:%s", method))
		}
	})
}

func TestTransformsCall(t *testing.T) {
	type TestResult struct {
		Method     string
		ResultHash string
		Args       []tengoLib.Object
		Error      error
	}
	noChangesHash := "a692a0f768855173"
	methods := []TestResult{
		TestResult{
			Method:     "resize",
			Args:       make([]tengoLib.Object, 0),
			Error:      tengoLib.ErrWrongNumArguments,
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "resize",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				tengoLib.FalseValue,
				tengoLib.FalseValue,
				tengoLib.FalseValue,
			},
			Error:      nil,
			ResultHash: "8bb55054d70af2be",
		},
		TestResult{
			Method: "resize",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				tengoLib.FalseValue,
				tengoLib.FalseValue,
			},
			Error:      nil,
			ResultHash: "16ef4f09495781bc",
		},
		TestResult{
			Method:     "extract",
			Args:       make([]tengoLib.Object, 0),
			Error:      tengoLib.ErrWrongNumArguments,
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "extract",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
			},
			Error:      nil,
			ResultHash: "26dc196780fc0611",
		},
		TestResult{
			Method: "extract",
			Args: []tengoLib.Object{
				&tengoLib.Map{Value: nil},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
			},
			Error:      tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: "map"},
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "crop",
			Args: []tengoLib.Object{
				&tengoLib.Map{Value: nil},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
			},
			Error:      tengoLib.ErrInvalidArgumentType{Name: "width", Expected: "int", Found: "map"},
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "crop",
			Args: []tengoLib.Object{
				&tengoLib.Map{Value: nil},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
			},
			Error:      tengoLib.ErrWrongNumArguments,
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "crop",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
				&tengoLib.String{Value: "gravity"},
				tengoLib.TrueValue,
				tengoLib.TrueValue,
			},
			Error:      nil,
			ResultHash: "d81a25214cb6d18f",
		},
		TestResult{
			Method: "resizeCropAuto",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
			},
			Error:      nil,
			ResultHash: "523065577321b2fa",
		},
		TestResult{
			Method: "resizeCropAuto",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
				&tengoLib.String{Value: "gravity"},
			},
			ResultHash: noChangesHash,
			Error:      tengoLib.ErrInvalidArgumentType{Name: "height", Expected: "int", Found: "string"},
		},
		TestResult{
			Method:     "interlace",
			Args:       []tengoLib.Object{},
			ResultHash: "f2ff5038353619c7",
			Error:      nil,
		},
		TestResult{
			Method:     "quality",
			Args:       []tengoLib.Object{},
			ResultHash: noChangesHash,
			Error:      tengoLib.ErrWrongNumArguments,
		},
		TestResult{
			Method: "quality",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
			},
			ResultHash: "ea20f91e50171468",
			Error:      nil,
		},
		TestResult{
			Method: "quality",
			Args: []tengoLib.Object{
				&tengoLib.String{Value: "gravity"},
			},
			ResultHash: noChangesHash,
			Error:      tengoLib.ErrInvalidArgumentType{Name: "quality", Expected: "int", Found: "string"},
		},
		TestResult{
			Method:     "stripMetadata",
			Args:       []tengoLib.Object{},
			ResultHash: "34ff3721dee2880c",
			Error:      nil,
		},
		TestResult{
			Method:     "grayscale",
			Args:       []tengoLib.Object{},
			ResultHash: "b2299a73127c840c",
			Error:      nil,
		},
		TestResult{
			Method: "rotate",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
				&tengoLib.Int{Value: 100},
			},
			Error:      tengoLib.ErrWrongNumArguments,
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "rotate",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
			},
			Error:      nil,
			ResultHash: "13a03cf3bd8c54e9",
		},
		TestResult{
			Method: "blur",
			Args: []tengoLib.Object{
				&tengoLib.Int{Value: 100},
			},
			Error:      tengoLib.ErrWrongNumArguments,
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "blur",
			Args: []tengoLib.Object{
				&tengoLib.Float{Value: 10.1},
				&tengoLib.Float{Value: 10.1},
			},
			Error:      nil,
			ResultHash: "d9e19da295225e46",
		},
		TestResult{
			Method: "blur",
			Args: []tengoLib.Object{
				&tengoLib.Float{Value: 10.1},
				&tengoLib.String{Value: "d"},
			},
			Error:      tengoLib.ErrInvalidArgumentType{Name: "minAmpl", Expected: "float64", Found: "string"},
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "format",
			Args: []tengoLib.Object{
				&tengoLib.String{Value: "png"},
			},
			Error:      nil,
			ResultHash: "e4ec06f5a498e779",
		},
		TestResult{
			Method:     "format",
			Args:       []tengoLib.Object{},
			Error:      tengoLib.ErrWrongNumArguments,
			ResultHash: noChangesHash,
		},
		TestResult{
			Method:     "watermark",
			Args:       []tengoLib.Object{},
			Error:      tengoLib.ErrWrongNumArguments,
			ResultHash: noChangesHash,
		},
		TestResult{
			Method: "watermark",
			Args: []tengoLib.Object{
				&tengoLib.String{Value: "image.png"},
				&tengoLib.String{Value: "top-left"},
				&tengoLib.Float{Value: 11.1},
			},
			Error:      nil,
			ResultHash: "12534f2e185c287e",
		},
	}

	for i, m := range methods {
		t.Run(fmt.Sprintf("method %s - %d", m.Method, i), func(t *testing.T) {
			c := transforms.Transforms{}
			tengoObject := tengo.Transforms{Value: c}
			v, err := tengoObject.IndexGet(&tengoLib.String{Value: m.Method})
			assert.Nil(t, err)
			_, err = v.Call(m.Args...)
			assert.Equal(t, err, m.Error)
			assert.Equal(t, m.ResultHash, tengoObject.String())

		})
	}
}
