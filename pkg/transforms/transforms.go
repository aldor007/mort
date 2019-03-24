package transforms

import (
	"encoding/binary"
	"errors"
	"hash"
	"strings"

	"github.com/aldor007/mort/pkg/helpers"
	"github.com/spaolacci/murmur3"
	"gopkg.in/h2non/bimg.v1"
	"math"
)

var watermarkPosX = map[string]float32{
	"left":   0,
	"center": 1. / 3.,
	"right":  2. / 3.,
}

var watermarkPosY = map[string]float32{
	"top":    0,
	"center": 1. / 3.,
	"bottom": 2. / 3.,
}

var cropGravity = map[string]bimg.Gravity{
	"center": bimg.GravityCentre,
	"north":  bimg.GravityNorth,
	"west":   bimg.GravityWest,
	"east":   bimg.GravityEast,
	"south":  bimg.GravitySouth,
	"smart":  bimg.GravitySmart,
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

var angleMap = map[int]bimg.Angle{
	0: bimg.D0,
	1: bimg.D90,
	2: bimg.D180,
	3: bimg.D270,
}
var prime64 = 1099511628211

func (w watermark) fetchImage() ([]byte, error) {
	return helpers.FetchObject(w.image)
}

func (w watermark) calculatePostion(width, height int) (top int, left int) {
	top = int(watermarkPosY[w.yPos] * float32(height))
	left = int(watermarkPosX[w.yPos] * float32(width))
	return
}

// ImageInfo holds information about image
type ImageInfo struct {
	width       int    // width of image in px
	height      int    // height of image in px
	format      string // format of image in string e.x. "jpg"
	orientation int
}

// NewImageInfo create new ImageInfo object from bimg metadata
func NewImageInfo(metadata bimg.ImageMetadata, format string) ImageInfo {
	return ImageInfo{width: metadata.Size.Width, height: metadata.Size.Height, format: format, orientation: metadata.Orientation}
}

// Transforms struct hold information about what operations should be performed on image
type Transforms struct {
	height         int
	width          int
	areaHeight     int
	areaWidth      int
	quality        int
	compression    int
	zoom           int
	crop           bool
	enlarge        bool
	embed          bool
	flip           bool
	flop           bool
	force          bool
	noAutoRotate   bool
	noProfile      bool
	interlace      bool
	stripMetadata  bool
	trim           bool
	rotate         bimg.Angle
	interpretation bimg.Interpretation
	gravity        bimg.Gravity
	blur           blur
	format         bimg.ImageType
	FormatStr      string

	watermark watermark

	NotEmpty bool
	NoMerge  bool

	autoCropWidth  int
	autoCropHeight int

	transHash fnvI64
}

func New() Transforms {
	t := Transforms{}
	return t
}

// Resize change image width and height
func (t *Transforms) Resize(width, height int, enlarge bool) error {
	t.width = width
	t.height = height
	t.enlarge = enlarge

	t.transHash.write(1111, uint64(t.width)*7, uint64(t.height)*3)

	if t.enlarge {
		t.transHash.write(12311)
	}

	t.NotEmpty = true
	return nil
}

// Crop extract part of image
func (t *Transforms) Crop(width, height int, gravity string, enlarge, embed bool) error {
	t.width = width
	t.height = height
	t.enlarge = enlarge
	t.crop = true
	t.embed = embed
	t.NotEmpty = true
	if g, ok := cropGravity[gravity]; ok {
		t.gravity = g
	} else {
		t.gravity = bimg.GravitySmart
	}

	t.transHash.write(1212, uint64(t.width)*5, uint64(t.height), uint64(t.gravity))
	if t.embed {
		t.transHash.write(3333)
	}

	if t.enlarge {
		t.transHash.write(54324)
	}

	return nil
}

// Crop extract part of image
func (t *Transforms) ResizeCropAuto(width, height int) error {
	t.NotEmpty = true
	t.autoCropWidth = width
	t.autoCropHeight = height
	t.NoMerge = true

	t.transHash.write(31229, uint64(width)*2, uint64(height))
	return nil
}

// Interlace enable image interlace
func (t *Transforms) Interlace() error {
	t.interlace = true
	t.NotEmpty = true
	t.transHash.write(1311, 71)
	return nil
}

// Quality change image quality
func (t *Transforms) Quality(quality int) error {
	t.quality = quality
	t.NotEmpty = true
	t.transHash.write(1401, uint64(t.quality))
	return nil
}

// StripMetadata remove EXIF from image
func (t *Transforms) StripMetadata() error {
	t.stripMetadata = true
	t.NotEmpty = true
	t.transHash.write(1999)
	return nil
}

// Blur blur whole image
func (t *Transforms) Blur(sigma, minAmpl float64) error {
	t.NotEmpty = true
	t.blur.sigma = sigma
	t.blur.minAmpl = minAmpl
	t.transHash.write(19121, uint64(t.blur.sigma*1000), uint64(t.blur.minAmpl*1000))
	return nil
}

// Hash return unique transform identifier
func (t *Transforms) Hash() hash.Hash64 {
	hashValue := murmur3.New64WithSeed(20171108)
	transHashB := make([]byte, 8)
	binary.LittleEndian.PutUint64(transHashB, t.transHash.value())
	hashValue.Write(transHashB)
	return hashValue
}

// Format change image format
func (t *Transforms) Format(format string) error {
	t.NotEmpty = true
	f, err := imageFormat(format)
	if err != nil {
		return err
	}
	t.format = f
	t.FormatStr = format
	t.transHash.write(1122121, uint64(f))
	return nil
}

// Watermark merge two image in one
func (t *Transforms) Watermark(image string, position string, opacity float32) error {
	if image == "" || position == "" {
		return errors.New("missing required params image or position")
	}

	p := strings.Split(position, "-")
	if len(p) != 2 {
		return errors.New("invalid position given")
	}

	if _, ok := watermarkPosY[p[0]]; !ok {
		return errors.New("invalid first position argument")
	}

	if _, ok := watermarkPosX[p[1]]; !ok {
		return errors.New("invalid second position argument")
	}

	if image == "" {
		return errors.New("empty image")
	}

	t.NotEmpty = true
	t.transHash.write(171200, uint64(len(image)), uint64(len(position)), uint64(opacity*100))
	t.watermark = watermark{image: image, xPos: p[1], yPos: p[0], opacity: opacity}
	return nil
}

// Grayscale convert image to B&W
func (t *Transforms) Grayscale() {
	t.interpretation = bimg.InterpretationBW
	t.transHash.write(32309)
	t.NotEmpty = true
}

// Rotate rotate image of given angle
func (t *Transforms) Rotate(angle int) error {
	a := int(angle / 90)
	if v, ok := angleMap[a]; ok {
		t.transHash.write(32941, uint64(a))
		t.rotate = v
		t.NotEmpty = true
		return nil
	}

	return errors.New("wrong angle")
}

// Merge append transformation from other object
func (t *Transforms) Merge(other Transforms) error {
	if other.NoMerge == true || t.NoMerge == true {
		return errors.New("unable to merge")
	}

	if other.NotEmpty == false {
		return nil
	}

	if other.watermark.image != "" {
		if t.watermark.image != "" {
			return errors.New("already have watermark")
		}
		t.watermark = other.watermark
	}

	if other.width != 0 {
		t.width = other.width
	}

	if other.height != 0 {
		t.height = other.height
	}

	if other.crop {
		t.crop = other.crop
	}

	if other.embed {
		t.embed = other.embed
	}

	t.autoCropWidth = other.autoCropWidth
	t.autoCropHeight = other.autoCropHeight

	if other.gravity != 0 {
		t.gravity = other.gravity
	}

	if other.blur.minAmpl != 0 {
		t.blur.minAmpl = t.blur.minAmpl + other.blur.minAmpl
	}

	if other.blur.sigma != 0 {
		t.blur.sigma = t.blur.sigma + other.blur.sigma
	}

	if other.interlace {
		t.interlace = other.interlace
	}

	if other.quality != 0 {
		t.quality = other.quality
	}

	if other.format != 0 {
		t.format = other.format
		t.FormatStr = other.FormatStr
	}

	if other.stripMetadata {
		t.stripMetadata = other.stripMetadata
	}

	t.transHash.write(other.transHash.value())
	t.NotEmpty = other.NotEmpty

	return nil
}

// Merge will merge tab of transformation into single one
func Merge(transformsTab []Transforms) []Transforms {
	transLen := len(transformsTab)
	if transLen <= 1 {
		return transformsTab
	}

	// revers order of transforms
	for i := 0; i < transLen/2; i++ {
		j := transLen - i - 1
		transformsTab[i], transformsTab[j] = transformsTab[j], transformsTab[i]
	}

	result := make([]Transforms, 1)
	baseTrans := transformsTab[0]
	result[0] = baseTrans
	for i := 1; i < transLen; i++ {
		if result[0].Merge(transformsTab[i]) != nil {
			result = append(result, transformsTab[i])
		}
	}

	return result
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

func (t *Transforms) calculateAutoCrop(info ImageInfo) (int, int, int, int) {
	if t.width != 0 {
		info.width = t.width
	}

	if t.height != 0 {
		info.height = t.height
	}

	wFactor := float64(info.width / t.autoCropWidth)
	hFactor := float64(info.height / t.autoCropHeight)

	var cropWidth, cropHeight float64
	if wFactor < hFactor {
		cropWidth = float64(info.width)
		cropHeight = math.Ceil(float64(t.autoCropHeight) * wFactor)
	} else {
		cropWidth = math.Ceil(float64(t.autoCropWidth) * hFactor)
		cropHeight = float64(info.height)
	}

	cropX := math.Floor((float64(info.width) - cropWidth) / 2.)
	cropY := math.Floor((float64(info.height) - cropHeight) / 5.)

	return int(cropX), int(cropY), int(cropWidth), int(cropHeight)
}

// BimgOptions return complete options for bimg lib
func (t *Transforms) BimgOptions(imageInfo ImageInfo) ([]bimg.Options, error) {
	var opts []bimg.Options
	b := bimg.Options{
		Width:         t.width,
		Height:        t.height,
		Enlarge:       t.enlarge,
		Crop:          t.crop,
		Embed:         t.embed,
		Interlace:     t.interlace,
		Quality:       t.quality,
		StripMetadata: t.stripMetadata,
		GaussianBlur: bimg.GaussianBlur{
			Sigma:   t.blur.sigma,
			MinAmpl: t.blur.minAmpl,
		},
		Rotate: t.rotate,
	}

	if t.gravity != 0 {
		b.Gravity = t.gravity
	}

	if t.FormatStr != "" {
		b.Type = t.format
	}

	if t.interpretation != 0 {
		b.Interpretation = t.interpretation
	}

	if t.watermark.image != "" {
		// fetch image
		buf, err := t.watermark.fetchImage()
		if err != nil {
			return opts, err
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

	opts = append(opts, b)

	if t.autoCropHeight != 0 || t.autoCropWidth != 0 {
		bAutoCrop := bimg.Options{}
		bAutoCrop.Left, bAutoCrop.Top, bAutoCrop.AreaWidth, bAutoCrop.AreaHeight = t.calculateAutoCrop(imageInfo)
		opts = append(opts, bAutoCrop)
		opts = append(opts, bimg.Options{Width: t.autoCropWidth, Height: t.autoCropHeight, Crop: true, Gravity: bimg.GravityCentre})
	}

	return opts, nil
}

//  FNV  for uint64
type fnvI64 uint64

func (f *fnvI64) write(data ...uint64) {
	hashValue := *f

	if hashValue == 0 {
		hashValue = fnvI64(1231)
	}

	for _, d := range data {
		hashValue ^= fnvI64(d)
		hashValue *= fnvI64(prime64)
	}

	*f = hashValue
}

func (f fnvI64) value() uint64 {
	return uint64(f)
}
