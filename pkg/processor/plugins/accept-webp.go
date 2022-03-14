package plugins

import (
	"net/http"
	"strings"

	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
)

func init() {
	RegisterPlugin("webp", WebpPlugin{})
}

// WebpPlugin plugins that transform image to webp if web browser can handle that format
type WebpPlugin struct {
}

func (WebpPlugin) configure(_ interface{}) {

}

// PreProcess add webp transform to object
func (WebpPlugin) preProcess(obj *object.FileObject, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "image/webp") && obj.HasTransform() {
		obj.Transforms.Format("webp")
		obj.AppendToKey("webp")
	}
}

// PostProcess update vary header
func (WebpPlugin) postProcess(obj *object.FileObject, req *http.Request, res *response.Response) {
	if res.IsImage() && obj.HasTransform() {
		res.Headers.Add("Vary", "Accept")
	}
}
