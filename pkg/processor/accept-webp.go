package processor

import (
	"context"
	"net/http"
	"strings"

	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
)

func init() {
	RegisterHook("webp", WebpHook{})
}

// WebpHook plugins that transform image to webp if web browser can handle that format
type WebpHook struct {
}

// preProcess add webp transform to object
func (_ WebpHook) preProcess(obj *object.FileObject, req *http.Request) {

	if strings.Contains(req.Header.Get("Accept"), "image/webp") && obj.HasTransform() {
		obj.Transforms.Format("webp")
		obj.UpdateKey("webp")
		ctx := obj.Ctx
		ctx = context.WithValue(ctx, "webp", true)
		obj.Ctx = ctx
	}
}

// postProcess update vary header
func (_ WebpHook) postProcess(obj *object.FileObject, req *http.Request, res *response.Response) {
	if res.IsImage() && obj.HasTransform() {
		res.Headers.Add("Vary", "Accept")
	}
}
