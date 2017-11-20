package transforms

import (
	"encoding/binary"
	"errors"
	"hash"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spaolacci/murmur3"
	"gopkg.in/h2non/bimg.v1"
)

var watermarkPosX = map[string]float32{
	"top":    0,
	"center": 1. / 3.,
	"bottom": 2. / 3.,
}

var watermarkPosY = map[string]float32{
	"left":   0,
	"center": 1. / 3.,
	"right":  2. / 3.,
}

type blur struct {
	sigma   float64
	minAmpl float64
}

type watermark struct {
	image   string
	opacity float32
	xPos    string
	yPos    string
}

func (w watermark) fetchImage() ([]byte, error) {
	if strings.HasPrefix(w.image, "http") {
		client := &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		req, err := http.NewRequest("GET", w.image, nil)
		if err != nil {
			return nil, err
		}

		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		defer response.Body.Close()
		buf, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		return buf, nil
	}

	f, err := os.Open(w.image)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return ioutil.ReadAll(f)
}

func (w watermark) calculatePostion(width, height int) (top int, left int) {
	top = int(watermarkPosY[w.yPos]*float32(height)) + 1
	left = int(watermarkPosX[w.yPos]*float32(width)) + 1
	return
}

// ImageInfo holds informaation about image
type ImageInfo struct {
	width  int    // width of image in px
	height int    // height of image in px
	format string // format of image in string e.x. "jpg"
}

// NewImageInfo create new ImageInfo object from bimg metadata
func NewImageInfo(metadata bimg.ImageMetadata, format string) ImageInfo {
	return ImageInfo{width: metadata.Size.Width, height: metadata.Size.Height, format: format}
}

// Transforms struct hold information about what operations should be performed on image
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

	blur blur

	format bimg.ImageType

	watermark watermark

	NotEmpty bool

	transHash uint64
}

// Resize change image width and height
func (t *Transforms) Resize(width, height int, enlarge bool) error {
	t.width = width
	t.height = height
	t.enlarge = enlarge

	t.transHash = 1000 + t.transHash + uint64(t.width) + uint64(t.height)

	if t.enlarge {
		t.transHash = t.transHash + 1
	}

	t.NotEmpty = true
	return nil
}

// Crop extract part of image
func (t *Transforms) Crop(width, height int, enlarge bool) error {
	t.width = width
	t.height = height
	t.enlarge = enlarge
	t.crop = true
	t.NotEmpty = true

	t.transHash = 1200 + t.transHash + uint64(t.width) + uint64(t.height)
	return nil
}

// Interlace enable image interlace
func (t *Transforms) Interlace() error {
	t.interlace = true
	t.NotEmpty = true
	t.transHash = 1300 + t.transHash + 71
	return nil
}

// Quality change image quality
func (t *Transforms) Quality(quality int) error {
	t.quality = quality
	t.NotEmpty = true
	t.transHash = 1400 + t.transHash + uint64(t.quality)
	return nil
}

// StripMetadata remove EXIF from image
func (t *Transforms) StripMetadata() error {
	t.stripMetadata = true
	t.NotEmpty = true
	t.transHash = 1500 + t.transHash + 85
	return nil
}

// Blur blur whole image
func (t *Transforms) Blur(sigma, minAmpl float64) error {
	t.NotEmpty = true
	t.blur.sigma = sigma
	t.blur.minAmpl = minAmpl
	t.transHash = 1600 + t.transHash + uint64(t.blur.sigma) + uint64(t.blur.minAmpl)
	return nil
}

// Hash return unique transform identifier
func (t *Transforms) Hash() hash.Hash64 {
	hash := murmur3.New64WithSeed(20171108)
	transHashB := make([]byte, 8)
	binary.LittleEndian.PutUint64(transHashB, t.transHash)
	hash.Write(transHashB)
	return hash
}

// Format change image format
func (t *Transforms) Format(format string) error {
	t.NotEmpty = true
	t.transHash = 1700 + t.transHash + 11
	f, err := imageFormat(format)
	if err != nil {
		return err
	}
	t.format = f
	return nil
}

// Watermark merge two image in one
func (t *Transforms) Watermark(image string, position string, opacity float32) error {
	if image == "" || position == "" {
		return errors.New("Missing required params")
	}
	p := strings.Split(position, "-")
	if len(p) != 2 {
		return errors.New("Invalid position given")
	}
	if _, ok := watermarkPosY[p[0]]; !ok {
		return errors.New("Invalid first position argument")
	}

	if _, ok := watermarkPosX[p[1]]; !ok {
		return errors.New("Invalid second position argument")
	}

	t.NotEmpty = true
	t.transHash = 1700 + t.transHash + uint64(len(image)) + uint64(len(position))
	t.watermark = watermark{image: image, xPos: p[1], yPos: p[0], opacity: opacity}
	return nil
}

func imageFormat(format string) (bimg.ImageType, error) {

	switch format {
	case "jpeg", "jpg":
		return bimg.JPEG, nil
	case "webp":
		return bimg.WEBP, nil
	case "png":
		return bimg.PNG, nil
	case "gif":
		return bimg.GIF, nil
	case "svg":
		return bimg.SVG, nil
	case "pdf":
		return bimg.PDF, nil
	default:
		return bimg.UNKNOWN, errors.New("Unknown format " + format)
	}
}

// BimgOptions return complete options for bimg lib
func (t *Transforms) BimgOptions(imageInfo ImageInfo) (bimg.Options, error) {
	b := bimg.Options{
		Width:         t.width,
		Height:        t.height,
		Enlarge:       t.enlarge,
		Crop:          t.crop,
		Interlace:     t.interlace,
		Quality:       t.quality,
		StripMetadata: t.stripMetadata,
		GaussianBlur: bimg.GaussianBlur{
			Sigma:   t.blur.sigma,
			MinAmpl: t.blur.minAmpl,
		},
	}

	if t.format != 0 {
		b.Type = t.format
	}

	if t.watermark.image != "" {
		// fetch image
		buf, err := t.watermark.fetchImage()
		if err != nil {
			return b, err
		}

		// calculate correct image dimensions
		width := imageInfo.width
		height := imageInfo.height

		if t.width != 0 && t.height != 0 {
			width = t.width
			height = t.height
		} else if t.width != 0 {
			width = t.width
			height = t.width * height / imageInfo.width
		} else if t.height != 0 {
			height = t.height
			width = t.height * width / imageInfo.height
		}

		top, left := t.watermark.calculatePostion(width, height)

		b.WatermarkImage = bimg.WatermarkImage{
			Left:    left,
			Top:     top,
			Buf:     buf,
			Opacity: t.watermark.opacity,
		}
	}

	return b, nil
}
