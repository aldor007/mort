package transforms

import "gopkg.in/h2non/bimg.v1"


type Transforms struct {
	Height         int
	Width          int
	AreaHeight     int
	AreaWidth      int
	Top            int
	Left           int
	Quality        int
	Compression    int
	Zoom           int
	Crop           bool
	Enlarge        bool
	Embed          bool
	Flip           bool
	Flop           bool
	Force          bool
	NoAutoRotate   bool
	NoProfile      bool
	Interlace      bool
	StripMetadata  bool
	Trim           bool

	NotEmpty       bool
}

func (self *Transforms) ResizeT(size []int, enlarge bool) (*Transforms) {
	self.Width = size[0]
	self.Height = size[1]
	self.Enlarge = enlarge
	self.NotEmpty = true
	return self
}

func (self *Transforms) CropT(size []int, enlarge bool) (*Transforms)  {
	self.Width = size[0]
	self.Height = size[1]
	self.Enlarge = enlarge
	self.Crop = true
	self.NotEmpty = true
	return self
}

func (self *Transforms) BimgOptions() (bimg.Options) {
	return bimg.Options{
		Width: self.Width,
		Height: self.Height,
		Enlarge: self.Enlarge,
		Crop: self.Crop,
	}
}


