package transforms

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/bimg.v1"
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

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
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

func TestTransformsResize(t *testing.T) {
	trans := Transforms{}
	trans.Resize(5, 100, false)

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Width, 5)
	assert.Equal(t, opts.Height, 100)
	assert.Equal(t, opts.Enlarge, false)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "755cf90e7c1d5f6", hashStr)

	trans2 := Transforms{}
	trans2.Resize(100, 5, false)

	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}

func TestTransformsCrop(t *testing.T) {
	trans := Transforms{}
	trans.Crop(11, 12, "smart", false)

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Width, 11)
	assert.Equal(t, opts.Height, 12)
	assert.Equal(t, opts.Enlarge, false)
	assert.Equal(t, opts.Gravity, bimg.GravitySmart)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "276ec67309d8039b", hashStr)

	trans2 := Transforms{}
	trans.Crop(12, 11, "smart", false)

	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}

func TestTransformsQuality(t *testing.T) {
	trans := Transforms{}
	trans.Quality(60)

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
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

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.True(t, opts.Interlace)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "f2ff5038353619c7", hashStr)
}

func TestTransformsStripMetadata(t *testing.T) {
	trans := Transforms{}
	trans.StripMetadata()

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.True(t, opts.StripMetadata)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "34ff3721dee2880c", hashStr)
}

func TestTransformsFormat(t *testing.T) {
	trans := Transforms{}
	trans.Format("jpeg")

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Type, bimg.JPEG)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "b34c4907588add8a", hashStr)

	trans2 := Transforms{}
	trans2.Format("png")
	hashStr2 := strconv.FormatUint(uint64(trans2.Hash().Sum64()), 16)
	assert.NotEqual(t, hashStr, hashStr2)
}

func TestTransformsGrayscale(t *testing.T) {
	trans := Transforms{}
	trans.Grayscale()

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)

	assert.Equal(t, opts.Interpretation, bimg.InterpretationBW)

	hashStr := strconv.FormatUint(uint64(trans.Hash().Sum64()), 16)
	assert.Equal(t, "b2299a73127c840c", hashStr)
}

func TestTransformsRotate(t *testing.T) {
	trans := Transforms{}
	trans.Rotate(90)

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
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

	opts, err := trans.BimgOptions(ImageInfo{})

	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.True(t, trans.NotEmpty)
}
