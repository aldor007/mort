package transforms

import (
	"github.com/h2non/bimg"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestNewImageInfo(t *testing.T) {
	metadata := bimg.ImageMetadata{}
	metadata.Size.Width = 100
	metadata.Size.Height = 150

	info := NewImageInfo(metadata, "jpg")
	assert.Equal(t, info.width, 100)
	assert.Equal(t, info.height, 150)
	assert.Equal(t, info.format, "jpg")
}

func TestTransformsBlur(t *testing.T) {
	trans := Transforms{}
	trans.Blur(1, 2)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]

	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.GaussianBlur.Sigma, 1.0)
	assert.Equal(t, opts.GaussianBlur.MinAmpl, 2.0)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "70427f089493f542", hashStr)

	trans2 := Transforms{}
	trans2.Blur(1, 2.1)

	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}
func TestTransformsJSON(t *testing.T) {
	trans := Transforms{}
	trans.Blur(1, 2)

	d := trans.ToJSON()

	assert.Equal(t, d["hash"], trans.HashStr())
}

func TestTransformsResize(t *testing.T) {
	trans := Transforms{}
	trans.Resize(5, 100, true, false, false)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Width, 5)
	assert.Equal(t, opts.Height, 100)
	assert.Equal(t, opts.Enlarge, true)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "3c9adb04ba75bd9c", hashStr)

	trans2 := Transforms{}
	trans2.Resize(100, 5, false, false, false)

	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}

func TestTransformsCrop(t *testing.T) {
	trans := Transforms{}
	trans.Crop(11, 12, "smart", false, false)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Width, 11)
	assert.Equal(t, opts.Height, 12)
	assert.Equal(t, opts.Enlarge, false)
	assert.Equal(t, opts.Gravity, bimg.GravitySmart)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "276ec67309d8039b", hashStr)

	trans2 := Transforms{}
	trans.Crop(12, 11, "unknown", false, false)

	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}

func TestTransformsQuality(t *testing.T) {
	trans := Transforms{}
	trans.Quality(60)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Quality, 60)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "b54b444cb2b62198", hashStr)

	trans2 := Transforms{}
	trans.Quality(61)

	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}

func TestTransformsInterlace(t *testing.T) {
	trans := Transforms{}
	trans.Interlace()

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.True(t, opts.Interlace)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "f2ff5038353619c7", hashStr)
}

func TestTransformsStripMetadata(t *testing.T) {
	trans := Transforms{}
	trans.StripMetadata()

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.True(t, opts.StripMetadata)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "34ff3721dee2880c", hashStr)
}

func TestTransformsFormat(t *testing.T) {
	trans := Transforms{}
	trans.Format("jpeg")

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Type, bimg.JPEG)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "b34c4907588add8a", hashStr)

	trans2 := Transforms{}
	trans2.Format("png")
	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)

	err = trans.Format("webp")

	assert.Nil(t, err)
	assert.Equal(t, trans.FormatStr, "webp")

	err = trans.Format("svg")

	assert.Nil(t, err)
	assert.Equal(t, trans.FormatStr, "svg")

	err = trans.Format("pdf")

	assert.Nil(t, err)
	assert.Equal(t, trans.FormatStr, "pdf")

	err = trans.Format("pdfaa")

	assert.NotNil(t, err)

}

func TestTransformsGrayscale(t *testing.T) {
	trans := Transforms{}
	trans.Grayscale()

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Interpretation, bimg.InterpretationBW)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "b2299a73127c840c", hashStr)
}

func TestTransformsRotate(t *testing.T) {
	trans := Transforms{}
	trans.Rotate(90)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Rotate, bimg.D90)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "13a03cf3bd8c54e9", hashStr)

	trans2 := Transforms{}
	trans2.Rotate(180)
	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}

func TestTransforms_Watermark(t *testing.T) {
	trans := Transforms{}
	trans.Watermark("../processor/benchmark/local/small.jpg", "top-left", 0.5)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	err = trans.Watermark("../processor/benchmark/local/small.jpg", "topleft", 0.5)

	assert.NotNil(t, err)

	err = trans.Watermark("../processor/benchmark/local/small.jpg", "left-top", 0.5)

	assert.NotNil(t, err)

	err = trans.Watermark("../processor/benchmark/local/small.jpg", "top-b", 0.5)

	assert.NotNil(t, err)

	err = trans.Watermark("", "top-left", 0.5)

	assert.NotNil(t, err)
}

func TestTransforms_Merge_Resize(t *testing.T) {
	tab := make([]Transforms, 2)
	tab[0].Resize(100, 0, false, false, false)

	tab[1].Resize(0, 300, true, false, false)

	result := Merge(tab)

	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].width, 100)
	assert.Equal(t, result[0].height, 300)
	assert.Equal(t, result[0].enlarge, true)
}

func TestTransforms_Merge_Crop(t *testing.T) {
	tab := make([]Transforms, 2)
	tab[0].Crop(4444, 0, "smart", false, false)

	tab[1].Crop(0, 120, "smart", false, true)

	result := Merge(tab)

	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].width, 4444)
	assert.Equal(t, result[0].height, 120)
	assert.Equal(t, result[0].enlarge, false)
	assert.Equal(t, result[0].embed, true)
	assert.Equal(t, result[0].crop, true)
	assert.Equal(t, result[0].gravity, bimg.GravitySmart)
}

func TestTransforms_Merge_Blur(t *testing.T) {
	tab := make([]Transforms, 3)
	tab[0].Blur(1., 3.)
	tab[1].Blur(2., 4.)
	tab[2].Blur(3., 3.)

	result := Merge(tab)

	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].blur.sigma, 6.)
	assert.Equal(t, result[0].blur.minAmpl, 10.)
}

func TestTransforms_Merge_Single(t *testing.T) {
	tab := make([]Transforms, 1)
	tab[0].Blur(1., 3.)

	result := Merge(tab)

	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].blur.sigma, 1.)
	assert.Equal(t, result[0].blur.minAmpl, 3.)
}

func TestTransforms_Merge_MultiTrans(t *testing.T) {
	tab := make([]Transforms, 4)
	tab[0].Blur(1., 3.)
	tab[0].Quality(10)
	tab[1].Interlace()
	tab[2].StripMetadata()
	tab[3].Format("webp")

	result := Merge(tab)

	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].blur.sigma, 1.)
	assert.Equal(t, result[0].blur.minAmpl, 3.)
	assert.Equal(t, result[0].interlace, true)
	assert.Equal(t, result[0].stripMetadata, true)
	assert.Equal(t, result[0].format, bimg.WEBP)
	assert.Equal(t, result[0].FormatStr, "webp")
	assert.Equal(t, result[0].quality, 10)
}

func TestTransforms_Merge_Empty(t *testing.T) {
	tab := make([]Transforms, 1)

	result := Merge(tab)

	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].NotEmpty, false)
}

func TestTransforms_Merge_Watermark(t *testing.T) {
	tab := make([]Transforms, 3)
	tab[0].Blur(1., 3.)
	tab[0].Watermark("image2", "top-left", 0.7)
	tab[1].Watermark("image", "top-left", 0.5)
	tab[2].Blur(3., 3.)

	result := Merge(tab)

	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0].blur.sigma, 3.)
	assert.Equal(t, result[0].blur.minAmpl, 3.)
	assert.Equal(t, result[1].watermark.image, "image2")
}

func TestTransforms_Extract(t *testing.T) {
	trans := Transforms{}
	trans.Extract(3, 1, 100, 200)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	opts := optsArr[0]
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Top, 3)
	assert.Equal(t, opts.Left, 1)
	assert.Equal(t, opts.AreaWidth, 100)
	assert.Equal(t, opts.AreaHeight, 200)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "76e63c06a9aacad3", hashStr)
}

func TestTransforms_ResizeCropAuto(t *testing.T) {
	trans := New()
	trans.ResizeCropAuto(210, 200)

	optsArr, err := trans.BimgOptions(ImageInfo{})
	assert.Nil(t, err)
	assert.Equal(t, len(optsArr), 3)
	assert.True(t, trans.NotEmpty)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "a9476be4baa3fb94", hashStr)
}

// Validation Tests

func TestQualityValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		quality int
		wantErr bool
	}{
		{"valid quality 1", 1, false},
		{"valid quality 50", 50, false},
		{"valid quality 100", 100, false},
		{"invalid quality 0", 0, true},
		{"invalid quality -1", -1, true},
		{"invalid quality 101", 101, true},
		{"invalid quality 200", 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := New()
			err := trans.Quality(tt.quality)
			if tt.wantErr {
				assert.NotNil(t, err, "Quality(%d) should return error", tt.quality)
				assert.Contains(t, err.Error(), "quality must be between 1 and 100")
			} else {
				assert.Nil(t, err, "Quality(%d) should not return error", tt.quality)
			}
		})
	}
}

func TestBlurValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sigma   float64
		minAmpl float64
		wantErr bool
	}{
		{"valid sigma 0.1", 0.1, 0, false},
		{"valid sigma 1.0", 1.0, 2.0, false},
		{"valid sigma 10.0", 10.0, 5.0, false},
		{"invalid sigma 0", 0, 0, true},
		{"invalid sigma -1", -1.0, 0, true},
		{"invalid sigma -5", -5.0, 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := New()
			err := trans.Blur(tt.sigma, tt.minAmpl)
			if tt.wantErr {
				assert.NotNil(t, err, "Blur(%f, %f) should return error", tt.sigma, tt.minAmpl)
				assert.Contains(t, err.Error(), "sigma must be positive")
			} else {
				assert.Nil(t, err, "Blur(%f, %f) should not return error", tt.sigma, tt.minAmpl)
			}
		})
	}
}

func TestWatermarkOpacityValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opacity float32
		wantErr bool
	}{
		{"valid opacity 0", 0.0, false},
		{"valid opacity 0.5", 0.5, false},
		{"valid opacity 1.0", 1.0, false},
		{"invalid opacity -0.1", -0.1, true},
		{"invalid opacity 1.1", 1.1, true},
		{"invalid opacity 2.0", 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := New()
			err := trans.Watermark("test.jpg", "top-left", tt.opacity)
			if tt.wantErr {
				assert.NotNil(t, err, "Watermark with opacity %f should return error", tt.opacity)
				assert.Contains(t, err.Error(), "opacity must be between 0 and 1")
			} else {
				assert.Nil(t, err, "Watermark with opacity %f should not return error", tt.opacity)
			}
		})
	}
}

func TestExtractValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		top     int
		left    int
		width   int
		height  int
		wantErr bool
	}{
		{"valid extract", 0, 0, 100, 100, false},
		{"valid extract with offset", 10, 20, 100, 100, false},
		{"invalid negative top", -1, 0, 100, 100, true},
		{"invalid negative left", 0, -1, 100, 100, true},
		{"invalid negative width", 0, 0, -1, 100, true},
		{"invalid negative height", 0, 0, 100, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := New()
			err := trans.Extract(tt.top, tt.left, tt.width, tt.height)
			if tt.wantErr {
				assert.NotNil(t, err, "Extract(%d, %d, %d, %d) should return error", tt.top, tt.left, tt.width, tt.height)
				assert.Contains(t, err.Error(), "extract coordinates cannot be negative")
			} else {
				assert.Nil(t, err, "Extract(%d, %d, %d, %d) should not return error", tt.top, tt.left, tt.width, tt.height)
			}
		})
	}
}

func TestResizeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"valid resize both", 100, 200, false},
		{"valid resize width only", 100, 0, false},
		{"valid resize height only", 0, 200, false},
		{"invalid negative width", -100, 200, true},
		{"invalid negative height", 100, -200, true},
		{"invalid both negative", -100, -200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := New()
			err := trans.Resize(tt.width, tt.height, false, false, false)
			if tt.wantErr {
				assert.NotNil(t, err, "Resize(%d, %d) should return error", tt.width, tt.height)
				assert.Contains(t, err.Error(), "width and height cannot be negative")
			} else {
				assert.Nil(t, err, "Resize(%d, %d) should not return error", tt.width, tt.height)
			}
		})
	}
}

func TestCropValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"valid crop", 100, 200, false},
		{"invalid negative width", -100, 200, true},
		{"invalid negative height", 100, -200, true},
		{"invalid both negative", -100, -200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := New()
			err := trans.Crop(tt.width, tt.height, "smart", false, false)
			if tt.wantErr {
				assert.NotNil(t, err, "Crop(%d, %d) should return error", tt.width, tt.height)
				assert.Contains(t, err.Error(), "width and height cannot be negative")
			} else {
				assert.Nil(t, err, "Crop(%d, %d) should not return error", tt.width, tt.height)
			}
		})
	}
}

func TestResizeCropAutoValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"valid resize crop auto", 100, 200, false},
		{"invalid negative width", -100, 200, true},
		{"invalid negative height", 100, -200, true},
		{"invalid both negative", -100, -200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := New()
			err := trans.ResizeCropAuto(tt.width, tt.height)
			if tt.wantErr {
				assert.NotNil(t, err, "ResizeCropAuto(%d, %d) should return error", tt.width, tt.height)
				assert.Contains(t, err.Error(), "width and height cannot be negative")
			} else {
				assert.Nil(t, err, "ResizeCropAuto(%d, %d) should not return error", tt.width, tt.height)
			}
		})
	}
}
