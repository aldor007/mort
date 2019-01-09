package plugins

import (
	"compress/gzip"
	"github.com/aldor007/mort/pkg/helpers"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	brEnc "github.com/google/brotli/go/cbrotli"
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

// CompressPlugin plugins that transform image to webp if web browser can handle that format
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
func (CompressPlugin) preProcess(obj *object.FileObject, req *http.Request) {

}

// PostProcess update vary header
func (c CompressPlugin) postProcess(obj *object.FileObject, req *http.Request, res *response.Response) {
	acceptEnc := req.Header.Get("Accept-Encoding")
	contentType := res.Headers.Get("Content-Type")
	if acceptEnc == "" || contentType == "" || helpers.IsRangeOrCondition(req) || (res.ContentLength < 1000 && res.ContentLength != -1) {
		return
	}

	if c.brotli.enabled && strings.Contains(acceptEnc, "br") {
		for _, supportedType := range c.brotli.types {
			if contentType == supportedType {
				res.Headers.Set("Content-Encoding", "br")
				res.Headers.Add("Vary", "Accept-Encoding")
				res.BodyTransformer(func(w io.Writer) io.WriteCloser {
					br := brEnc.NewWriter(w, brEnc.WriterOptions{Quality: c.brotli.level})
					return br
				})
				return
			}
		}

	}

	if c.gzip.enabled && strings.Contains(acceptEnc, "gzip") {
		for _, supportedType := range c.gzip.types {
			if contentType == supportedType {
				res.Headers.Set("Content-Encoding", "gzip")
				res.Headers.Add("Vary", "Accept-Encoding")
				res.BodyTransformer(func(w io.Writer) io.WriteCloser {
					gzipW, err := gzip.NewWriterLevel(w, c.gzip.level)
					if err != nil {
						panic(err)
					}

					return gzipW
				})
				return
			}
		}
	}

}
