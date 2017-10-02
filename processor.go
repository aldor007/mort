package mort

import (
	"strings"

	"mort/config"
	"mort/engine"
	"mort/object"
	"mort/response"
	"mort/storage"
	"mort/transforms"
)

func Process(obj *object.FileObject) *response.Response {
	// first check if object is on storage
	res := updateHeaders(storage.Get(obj))

	if res.StatusCode != 404 {
		return res
	}

	// if not check if we can try to perform transformation on it

	// object doesn't have parent and we cannot do transfrom if it hasn't parent.
	// So we are returning response from storage
	if !obj.HasParent() {
		return res
	}

	var currObj *object.FileObject = obj
	var parentObj *object.FileObject
	var transforms []transforms.Transforms
	// search for last parent
	for currObj.HasParent() {
		if currObj.HasTransform() {
			transforms = append(transforms, currObj.Transforms)
		}
		currObj = currObj.Parent

		if !currObj.HasParent() {
			parentObj = currObj
		}
	}

	// get parent from storage
	res = updateHeaders(storage.Get(parentObj))

	if res.StatusCode == 404 {
		return res
	}

	if strings.Contains(res.Headers[response.ContentType], "image/") {
		// revers order of transforms
		for i := 0; i < len(transforms)/2; i++ {
			j := len(transforms) - i - 1
			transforms[i], transforms[j] = transforms[j], transforms[i]
		}

		return updateHeaders(processImage(res, transforms))
	}

	return updateHeaders(storage.Get(obj))
}

func processImage(parent *response.Response, transforms []transforms.Transforms) *response.Response {
	engine := engine.NewImageEngine(parent)
	result, err := engine.Process(transforms)
	if err != nil {
		return response.NewError(400, err)
	}

	return result

}

func updateHeaders(res *response.Response) *response.Response {
	headers := config.GetInstance().Headers
	for _, headerPred := range headers {
		for _, status := range headerPred.StatusCodes {
			if status == res.StatusCode {
				for h, v := range headerPred.Values {
					res.Set(h, v)
				}
				return res
			}
		}
	}
	return res
}
