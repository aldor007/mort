package plugins

import (
	"compress/gzip"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"io"
	"net/http"
	"strings"
)

func init() {
	RegisterPlugin("compress", &CompressPlugin{})
}

type compressConfig struct {
	level   int
	types   []string
	enabled bool
}

// WebpPlugin plugins that transform image to webp if web browser can handle that format
type CompressPlugin struct {
	brotli compressConfig
	gzip   compressConfig
}

func parseConfig(cType *compressConfig, cfg interface{}) {
	cfgKeys := cfg.(map[interface{}]interface{})

	if types, ok := cfgKeys["types"]; ok {
		typesArr := types.([]interface{})
		for _, t := range typesArr {
			cType.types = append(cType.types, t.(string))
		}
	} else {
		cType.types = []string{"text/html"}
	}

	if cLevel, ok := cfgKeys["level"]; ok {
		cType.level = cLevel.(int)
	} else {
		cType.level = 4
	}

	cType.enabled = true
}

func (c *CompressPlugin) configure(config interface{}) {
	cfg := config.(map[interface{}]interface{})

	if tmpCfg, ok := cfg["brotli"]; ok {
		parseConfig(&c.brotli, tmpCfg)
	}

	if tmpCfg, ok := cfg["gzip"]; ok {
		parseConfig(&c.gzip, tmpCfg)
	}
}

// PreProcess add webp transform to object
func (_ CompressPlugin) preProcess(obj *object.FileObject, req *http.Request) {

}

// PostProcess update vary header
func (c CompressPlugin) postProcess(obj *object.FileObject, req *http.Request, res *response.Response) {
	acceptEnc := req.Header.Get("Accept-Encoding")
	contentType := res.Headers.Get("Content-Type")
	if acceptEnc == "" || contentType == "" {
		return
	}

	if c.gzip.enabled && strings.Contains(acceptEnc, "gzip") {
		res.Headers.Set("Content-Encoding", "gzip")
		res.Headers.Add("Vary", "Accept-Encoding")
		for _, supportedType := range c.gzip.types {
			if contentType == supportedType {
				res.BodyTransformer(func(w io.Writer) io.WriteCloser {
					gzipW := gzip.NewWriter(w)
					return gzipW
				})
				return
			}
		}
	}

}
