package transforms

import (
	"gopkg.in/h2non/bimg.v1"
)

type Transforms struct {
	height        int
	width         int
	areaHeight    int
	areaWidth     int
	top           int
	left          int
	quality       int
	compression   int
	zoom          int
	crop          bool
	enlarge       bool
	embed         bool
	flip          bool
	flop          bool
	force         bool
	noAutoRotate  bool
	noProfile     bool
	interlace     bool
	stripMetadata bool
	trim          bool

	sigma         float64
	minAmpl       float64

	NotEmpty bool
}

func (t *Transforms) Resize(size []int, enlarge bool) *Transforms {
	t.width = size[0]
	if len(size) == 2 {
		t.height = size[1]
	}
	t.enlarge = enlarge
	t.NotEmpty = true
	return t
}

func (t *Transforms) Crop(size []int, enlarge bool) *Transforms {
	t.width = size[0]
	t.height = size[1]
	t.enlarge = enlarge
	t.crop = true
	t.NotEmpty = true
	return t
}

func (t *Transforms) Interlace()  *Transforms{
	t.interlace = true
	t.NotEmpty = true
	return t
}

func (t *Transforms) Quality(quality int) *Transforms {
	t.quality = quality
	t.NotEmpty = true
	return t
}

func (t *Transforms) StripMetadata() *Transforms {
	t.stripMetadata = true
	t.NotEmpty = true
	return t
}

func (t *Transforms) Blur(sigma, minAmpl float64) *Transforms {
	t.NotEmpty = true
	t.sigma = sigma
	t.minAmpl = minAmpl
	return t
}

func (t *Transforms) BimgOptions() bimg.Options {
	return bimg.Options{
		Width:   t.width,
		Height:  t.height,
		Enlarge: t.enlarge,
		Crop:    t.crop,
		Interlace: t.interlace,
		Quality: t.quality,
		StripMetadata: t.stripMetadata,
		GaussianBlur: bimg.GaussianBlur{
			Sigma: t.sigma,
			MinAmpl: t.minAmpl,
		},
	}
}
