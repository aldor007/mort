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

func TestTransformsResize(t *testing.T) {
	trans := Transforms{}
	trans.Resize(5, 100, true)

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
	trans2.Resize(100, 5, false)

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
	tab[0].Resize(100, 0, false)

	tab[1].Resize(0, 300, true)

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
	tab[0].Watermark("image2", "top-left", 2.)
	tab[1].Watermark("image", "top-left", 2.)
	tab[2].Blur(3., 3.)

	result := Merge(tab)

	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0].blur.sigma, 3.)
	assert.Equal(t, result[0].blur.minAmpl, 3.)
	assert.Equal(t, result[1].watermark.image, "image2")
}
