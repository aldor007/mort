package transforms

type Base interface{


}

type ICrop struct {
	Size []int  `json:"size"`
	Mode string `json:"mode"`
}

type Thumbnail struct {
	ICrop
}

type Crop struct {
	ICrop
	Start []int `json:"start"`
}

type SmartCrop struct {
	ICrop
}
