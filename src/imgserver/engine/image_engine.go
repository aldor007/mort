package engine

import "imgserver/transforms"

type ImageEngine struct {
	Input byte[]
	Output byte[]
}


func NewImageEngine (body byte[],  result &byte[]) ImageEngine {
	return &IamgeEngine{Input: body, Output: result}
}

func (self *ImageEngine) Process(base transforms.Base) ImageEngine*{

}