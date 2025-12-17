package plugins

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/aldor007/mort/pkg/helpers"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	brEnc "github.com/google/brotli/go/cbrotli"
	"github.com/klauspost/compress/zstd"
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
	zstd   compressConfig
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

	if tmpCfg, ok := cfg["zstd"]; ok {
		parseConfig(&c.zstd, tmpCfg)
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
			if strings.Contains(contentType, supportedType) {
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

	if c.zstd.enabled && strings.Contains(acceptEnc, "zstd") {
		for _, supportedType := range c.zstd.types {
			if strings.Contains(contentType, supportedType) {
				res.Headers.Set("Content-Encoding", "zstd")
				res.Headers.Add("Vary", "Accept-Encoding")
				res.BodyTransformer(func(w io.Writer) io.WriteCloser {
					// Map compression level (1-10) to zstd level
					// zstd levels go from 1 (fastest) to 19 (best compression), default is 3
					zstdLevel := zstd.SpeedDefault
					if c.zstd.level >= 1 && c.zstd.level <= 4 {
						zstdLevel = zstd.SpeedFastest
					} else if c.zstd.level >= 5 && c.zstd.level <= 7 {
						zstdLevel = zstd.SpeedDefault
					} else if c.zstd.level >= 8 {
						zstdLevel = zstd.SpeedBestCompression
					}

					zw, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstdLevel))
					if err != nil {
						panic(err)
					}
					return zw
				})
				return
			}
		}
	}

	if c.gzip.enabled && strings.Contains(acceptEnc, "gzip") {
		for _, supportedType := range c.gzip.types {
			if strings.Contains(contentType, supportedType) {
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
