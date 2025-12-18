package engine

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/aldor007/mort/pkg/transforms"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestImageEngine_Process_Error(t *testing.T) {
	image := response.NewNoContent(500)
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, err := object.NewFileObjectFromPath("/local/small/parent.jpg", &mortConfig)

	assert.Nil(t, err)
	assert.NotNil(t, obj)

	e := NewImageEngine(image)
	res, err := e.Process(obj, []transforms.Transforms{obj.Transforms})

	assert.NotNil(t, err)
	assert.Equal(t, res.StatusCode, 500)
}

func TestImageEngine_Process(t *testing.T) {
	f, err := os.Open("testdata/small.jpg")
	if err != nil {
		panic(err)
	}

	image := response.New(200, f)
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, err := object.NewFileObjectFromPath("/local/parent.jpg?width=100&height=70", &mortConfig)

	assert.Nil(t, err)
	assert.NotNil(t, obj)

	obj.Transforms.Resize(100, 70, false, false, false)

	e := NewImageEngine(image)
	res, err := e.Process(obj, []transforms.Transforms{obj.Transforms})

	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("content-type"), "image/jpeg")
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "100")
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-height"), "70")
}

func TestImageEngine_Process_MultipleTransforms(t *testing.T) {
	// Note: Parent test doesn't use t.Parallel() because subtests do.
	// This avoids overwhelming the CI test coordinator.

	tests := []struct {
		name           string
		transformSetup func() []transforms.Transforms
		expectedWidth  string
		expectedHeight string
		expectedFormat string
		description    string
	}{
		{
			name: "should apply resize then crop",
			transformSetup: func() []transforms.Transforms {
				trans1 := transforms.Transforms{}
				trans1.Resize(200, 200, false, false, false)
				trans2 := transforms.Transforms{}
				trans2.Crop(100, 100, "center", false, false)
				return []transforms.Transforms{trans1, trans2}
			},
			expectedWidth:  "100",
			expectedHeight: "100",
			expectedFormat: "image/jpeg",
			description:    "resize followed by crop should produce 100x100 image",
		},
		{
			name: "should apply multiple sequential transforms",
			transformSetup: func() []transforms.Transforms {
				trans1 := transforms.Transforms{}
				trans1.Resize(150, 150, false, false, false)
				trans2 := transforms.Transforms{}
				trans2.Quality(80)
				trans3 := transforms.Transforms{}
				trans3.Interlace()
				return []transforms.Transforms{trans1, trans2, trans3}
			},
			expectedWidth:  "150",
			expectedHeight: "150",
			expectedFormat: "image/jpeg",
			description:    "multiple transforms should be applied in sequence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.Open("testdata/small.jpg")
			assert.Nil(t, err)

			image := response.New(200, f)
			mortConfig := config.Config{}
			mortConfig.Load("testdata/config.yml")
			obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
			assert.Nil(t, err)

			e := NewImageEngine(image)
			res, err := e.Process(obj, tt.transformSetup())

			assert.Nil(t, err, tt.description)
			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, tt.expectedFormat, res.Headers.Get("content-type"))
			assert.Equal(t, tt.expectedWidth, res.Headers.Get("x-amz-meta-public-width"))
			assert.Equal(t, tt.expectedHeight, res.Headers.Get("x-amz-meta-public-height"))
		})
	}
}

func TestImageEngine_Process_FormatConversion(t *testing.T) {
	// Note: Parent test doesn't use t.Parallel() because subtests do.
	// This avoids overwhelming the CI test coordinator.

	tests := []struct {
		name         string
		format       string
		expectedType string
		description  string
	}{
		{
			name:         "should convert to PNG",
			format:       "png",
			expectedType: "image/png",
			description:  "JPEG to PNG conversion",
		},
		{
			name:         "should convert to WebP",
			format:       "webp",
			expectedType: "image/webp",
			description:  "JPEG to WebP conversion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.Open("testdata/small.jpg")
			assert.Nil(t, err)

			image := response.New(200, f)
			mortConfig := config.Config{}
			mortConfig.Load("testdata/config.yml")
			obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
			assert.Nil(t, err)

			trans := transforms.Transforms{}
			trans.Format(tt.format)

			e := NewImageEngine(image)
			res, err := e.Process(obj, []transforms.Transforms{trans})

			assert.Nil(t, err, tt.description)
			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, tt.expectedType, res.Headers.Get("content-type"))
		})
	}
}

func TestImageEngine_Process_ETagGeneration(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/small.jpg")
	assert.Nil(t, err)

	image := response.New(200, f)
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, err := object.NewFileObjectFromPath("/local/parent.jpg?width=100", &mortConfig)
	assert.Nil(t, err)

	trans := transforms.Transforms{}
	trans.Resize(100, 0, false, false, false)

	e := NewImageEngine(image)
	res, err := e.Process(obj, []transforms.Transforms{trans})

	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	// Should have ETag header
	etag := res.Headers.Get("ETag")
	assert.NotEmpty(t, etag, "ETag should be generated")
	assert.Equal(t, 32, len(etag), "ETag should be MD5 hash (32 chars)")

	// Should have Last-Modified header
	lastModified := res.Headers.Get("Last-Modified")
	assert.NotEmpty(t, lastModified, "Last-Modified should be set")
}

func TestImageEngine_Process_CropOperations(t *testing.T) {
	// Note: Parent test doesn't use t.Parallel() because subtests do.
	// This avoids overwhelming the CI test coordinator.

	tests := []struct{
		name        string
		width       int
		height      int
		gravity     string
		description string
	}{
		{
			name:        "should crop with center gravity",
			width:       50,
			height:      50,
			gravity:     "center",
			description: "center crop should produce 50x50",
		},
		{
			name:        "should crop with north gravity",
			width:       60,
			height:      40,
			gravity:     "north",
			description: "north crop should align to top",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.Open("testdata/small.jpg")
			assert.Nil(t, err)

			image := response.New(200, f)
			mortConfig := config.Config{}
			mortConfig.Load("testdata/config.yml")
			obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
			assert.Nil(t, err)

			trans := transforms.Transforms{}
			trans.Crop(tt.width, tt.height, tt.gravity, false, false)

			e := NewImageEngine(image)
			res, err := e.Process(obj, []transforms.Transforms{trans})

			assert.Nil(t, err, tt.description)
			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, "image/jpeg", res.Headers.Get("content-type"))
		})
	}
}

func TestImageEngine_Process_QualityAndInterlace(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/small.jpg")
	assert.Nil(t, err)

	image := response.New(200, f)
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
	assert.Nil(t, err)

	trans := transforms.Transforms{}
	trans.Quality(50)
	trans.Interlace()

	e := NewImageEngine(image)
	res, err := e.Process(obj, []transforms.Transforms{trans})

	assert.Nil(t, err, "quality and interlace should be applied")
	assert.Equal(t, 200, res.StatusCode)

	// Verify image was processed
	body, err := res.Body()
	assert.Nil(t, err)
	assert.NotEmpty(t, body, "processed image should have content")
}

func TestImageEngine_Process_Grayscale(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/small.jpg")
	assert.Nil(t, err)

	image := response.New(200, f)
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
	assert.Nil(t, err)

	trans := transforms.Transforms{}
	trans.Grayscale()

	e := NewImageEngine(image)
	res, err := e.Process(obj, []transforms.Transforms{trans})

	assert.Nil(t, err, "grayscale should be applied")
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "image/jpeg", res.Headers.Get("content-type"))
}

func TestImageEngine_Process_Blur(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/small.jpg")
	assert.Nil(t, err)

	image := response.New(200, f)
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
	assert.Nil(t, err)

	trans := transforms.Transforms{}
	trans.Blur(10, 5)

	e := NewImageEngine(image)
	res, err := e.Process(obj, []transforms.Transforms{trans})

	assert.Nil(t, err, "blur should be applied")
	assert.Equal(t, 200, res.StatusCode)
}

func TestImageEngine_Process_Rotate(t *testing.T) {
	// Note: Parent test doesn't use t.Parallel() because subtests do.
	// This avoids overwhelming the CI test coordinator.

	tests := []struct {
		name        string
		angle       int
		description string
	}{
		{
			name:        "should rotate 90 degrees",
			angle:       90,
			description: "90 degree rotation",
		},
		{
			name:        "should rotate 180 degrees",
			angle:       180,
			description: "180 degree rotation",
		},
		{
			name:        "should rotate 270 degrees",
			angle:       270,
			description: "270 degree rotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.Open("testdata/small.jpg")
			assert.Nil(t, err)

			image := response.New(200, f)
			mortConfig := config.Config{}
			mortConfig.Load("testdata/config.yml")
			obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
			assert.Nil(t, err)

			trans := transforms.Transforms{}
			trans.Rotate(tt.angle)

			e := NewImageEngine(image)
			res, err := e.Process(obj, []transforms.Transforms{trans})

			assert.Nil(t, err, tt.description)
			assert.Equal(t, 200, res.StatusCode)
		})
	}
}

func TestImageEngine_Process_MetadataHeaders(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/small.jpg")
	assert.Nil(t, err)

	image := response.New(200, f)
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, err := object.NewFileObjectFromPath("/local/parent.jpg", &mortConfig)
	assert.Nil(t, err)

	trans := transforms.Transforms{}
	trans.Resize(120, 80, false, false, false)

	e := NewImageEngine(image)
	res, err := e.Process(obj, []transforms.Transforms{trans})

	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	// Verify all metadata headers are set
	assert.NotEmpty(t, res.Headers.Get("content-type"), "should have content-type")
	assert.NotEmpty(t, res.Headers.Get("ETag"), "should have ETag")
	assert.NotEmpty(t, res.Headers.Get("Last-Modified"), "should have Last-Modified")
	assert.NotEmpty(t, res.Headers.Get("x-amz-meta-public-width"), "should have width metadata")
	assert.NotEmpty(t, res.Headers.Get("x-amz-meta-public-height"), "should have height metadata")
	assert.Equal(t, "120", res.Headers.Get("x-amz-meta-public-width"))
	assert.Equal(t, "80", res.Headers.Get("x-amz-meta-public-height"))
}

func TestImageEngine_NewImageEngine(t *testing.T) {
	t.Parallel()

	res := response.NewString(200, "test")
	engine := NewImageEngine(res)

	assert.NotNil(t, engine, "should create new image engine")
	assert.NotNil(t, engine.parent, "should store parent response")
}
