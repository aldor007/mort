package mort

import (
	"errors"
	"strings"

	"mort/config"
	"mort/engine"
	"mort/object"
	"mort/response"
	"mort/storage"
	"mort/transforms"
	"github.com/labstack/echo"
)

func Process(ctx echo.Context, obj *object.FileObject) *response.Response {
	switch ctx.Request().Method {
		case "GET":
			return hanldeGET(obj)
		case "PUT":
			return handlePUT(ctx, obj)

	default:
		return response.NewError(405, errors.New("method not allowed"))
	}
}

func handlePUT(ctx echo.Context, obj *object.FileObject) *response.Response {
	return storage.Set(obj, ctx.Request().Header, ctx.Request().ContentLength, ctx.Request().Body)
}

func hanldeGET(obj *object.FileObject) *response.Response {
	var currObj *object.FileObject = obj
	var parentObj *object.FileObject = nil
	var transforms []transforms.Transforms
	var res        *response.Response
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
	if parentObj != nil {
		res = updateHeaders(storage.Get(parentObj))
	}

	if res.StatusCode != 200 {
		return res
	}

	// check if object is on storage
	res = updateHeaders(storage.Get(obj))
	if res.StatusCode != 404 {
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
