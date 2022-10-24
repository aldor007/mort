package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/aldor007/mort/pkg/helpers"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/transforms"
	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/zap"

	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	// HeaderContentType name of Content-Type header
	HeaderContentType = "content-type"
)

type bodyTransformFnc func(writer io.Writer) io.WriteCloser

// Response is helper struct for wrapping different storage response
type Response struct {
	StatusCode    int         // status code of response
	Headers       http.Header // headers for response
	ContentLength int64       // if buffered response contains length of buffer, for streams it equal to -1
	debug         bool        // debug flag
	errorValue    error       // error value

	reader     io.ReadCloser // reader for response body
	body       []byte        // response body for buffered value
	bodyReader io.ReadCloser // original response buffer
	bodySeeker io.ReadSeeker

	transformer bodyTransformFnc // function that can transform body writner
	cachable    bool             // flag indicating if response can be cached
	ttl         int              // time to live in cache
	trans       []transforms.Transforms
}

// New create response object with io.ReadCloser
func New(statusCode int, body io.ReadCloser) *Response {
	res := Response{StatusCode: statusCode, reader: body}
	res.ContentLength = 0
	if body != nil {
		seeker, ok := body.(io.ReadSeeker)
		if ok {
			res.bodySeeker = seeker
		}
		res.ContentLength = -1
	}
	res.Headers = make(http.Header)

	return &res
}

// NewNoContent create response object without content
func NewNoContent(statusCode int) *Response {
	res := New(statusCode, nil)
	res.ContentLength = 0
	return res
}

// NewString create response object from string
func NewString(statusCode int, body string) *Response {
	res := Response{StatusCode: statusCode}
	res.setBodyBytes([]byte(body))
	res.Headers = make(http.Header)
	res.Headers.Set(HeaderContentType, "text/plain")
	return &res
}

// NewBuf create response object from []byte
func NewBuf(statusCode int, body []byte) *Response {
	res := Response{StatusCode: statusCode}
	res.Headers = make(http.Header)
	res.setBodyBytes(body)
	return &res
}

// NewError create response object from error
func NewError(statusCode int, err error) *Response {
	res := Response{StatusCode: statusCode, errorValue: err}
	res.Headers = make(http.Header)
	res.Headers.Set(HeaderContentType, "application/json")
	res.setBodyBytes([]byte(`{"message": "error"}`))
	return &res
}

// SetContentType update content type header of response
func (r *Response) SetContentType(contentType string) *Response {
	r.Headers.Set(HeaderContentType, contentType)
	return r
}

// Set update response headers
func (r *Response) Set(headerName string, headerValue string) {
	r.Headers.Set(headerName, headerValue)
}

func (r *Response) setBodyBytes(body []byte) {
	if r.reader != nil {
		panic("reader must not be set when setBodyBytes is used")
	}
	r.bodySeeker = bytes.NewReader(body)
	r.reader = io.NopCloser(r.bodySeeker)
	r.ContentLength = int64(len(body))
	r.body = body
}

// Body reads all content of response and returns []byte
// Content of the response is changed
// Such response shouldn't be Send to client
func (r *Response) Body() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}

	if r.reader == nil {
		return nil, errors.New("empty body")
	}

	body, err := ioutil.ReadAll(r.reader)
	r.reader.Close()
	r.reader = nil
	r.setBodyBytes(body)
	return r.body, err
}

// CopyBody returns a copy of Body in []byte
func (r *Response) CopyBody() ([]byte, error) {
	var err error
	src := r.body
	if src == nil {
		src, err = r.Body()
		if err != nil {
			return nil, err
		}
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst, nil
}

// Close response reader
func (r *Response) Close() {
	if r.reader != nil {
		io.ReadAll(r.reader)
		r.reader.Close()
		r.reader = nil
	}

	if r.bodyReader != nil {
		io.ReadAll(r.bodyReader)
		r.bodyReader.Close()
		r.bodyReader = nil
	}
}

// SetDebug set flag indicating that response can including debug information
func (r *Response) SetDebug(obj *object.FileObject) *Response {
	if obj.Debug == true {
		r.debug = true
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("x-mort-key", obj.Key)
		r.Headers.Set("x-mort-storage", obj.Storage.Kind)

		if obj.HasTransform() {
			r.Headers.Set("x-mort-transform", "true")
			jsonRes := make([]map[string]interface{}, len(r.trans))
			for i, t := range r.trans {
				jsonRes[i] = t.ToJSON()
			}
			buf, err := json.Marshal(jsonRes)
			if err == nil {
				r.Headers.Set("x-mort-transform-json", string(buf))
			} else {
				monitoring.Log().Warn("Response/SetDebug unable to marshal trans", zap.Error(err))
			}
		}

		if obj.HasParent() {
			r.Headers.Set("x-mort-parent-key", obj.Parent.Key)
			r.Headers.Set("x-mort-parent-bucket", obj.Parent.Bucket)
			r.Headers.Set("x-mort-parent-storage", obj.Parent.Storage.Kind)
		}
		r.writeDebug()
		return r
	}

	r.debug = false
	return r
}

func (r *Response) SetTransforms(trans []transforms.Transforms) {
	r.trans = trans
}

// HasError check if response contains error
func (r *Response) HasError() bool {
	return r.errorValue != nil
}

// Error returns error instance
func (r *Response) Error() error {
	return r.errorValue
}

// Send write response to client using streaming
func (r *Response) Send(w http.ResponseWriter) error {
	for headerName, headerValue := range r.Headers {
		w.Header().Set(headerName, headerValue[0])
	}

	defer r.Close()

	var resStream io.ReadCloser
	if r.ContentLength == 0 {
		w.WriteHeader(r.StatusCode)
		return nil
	}

	if r.ContentLength > 0 && r.transformer == nil {
		w.Header().Set("content-length", strconv.FormatInt(r.ContentLength, 10))
	}

	w.WriteHeader(r.StatusCode)
	resStream = r.Stream()
	if resStream == nil {
		return nil
	}
	if r.transformer != nil {
		tW := r.transformer(w)
		io.Copy(tW, resStream)
		tW.Close()
	} else {
		io.Copy(w, resStream)
	}
	return resStream.Close()
}

// SendContent use http.ServeContent to return response to client
// It can handle range and condition requests
// In this function we don't need to use transformer because it don't serve whole body
// It is used for range and condition requests
func (r *Response) SendContent(req *http.Request, w http.ResponseWriter) error {
	// ServerContent will modified status code so to it we should pass only 200 response
	if r.StatusCode != 200 || r.bodySeeker == nil || helpers.IsRangeOrCondition(req) == false {
		return r.Send(w)
	}

	defer r.Close()
	for headerName, headerValue := range r.Headers {
		w.Header().Set(headerName, headerValue[0])
	}

	lastMod, err := time.Parse(http.TimeFormat, r.Headers.Get("Last-Modified"))
	if err != nil {
		lastMod = time.Now()
	}

	http.ServeContent(w, req, "", lastMod, r.bodySeeker)
	return nil
}

func (r *Response) IsCacheable() bool {
	r.parseCacheHeaders()
	return r.StatusCode > 199 && r.StatusCode < 299 && r.cachable
}

func (r *Response) IsFromCache() bool {
	return r.Headers.Get("x-mort-cache") != ""
}

func (r *Response) SetCacheHit() {
	r.Headers.Set("x-mort-cache", "hit")
}

func (r *Response) GetTTL() int {
	r.parseCacheHeaders()
	return r.ttl
}

func (r *Response) parseCacheHeaders() {
	reqDir, err := cacheobject.ParseRequestCacheControl(r.Headers.Get("Cache-Control"))
	if err != nil {
		r.cachable = false
		return
	}

	r.ttl = int(reqDir.MaxAge)
	if r.ttl > 0 {
		r.cachable = true
	}
}
func (r *Response) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(r.StatusCode, r.Headers, r.ttl, r.ContentLength, r.body, r.cachable)
}

func (r *Response) DecodeMsgpack(dec *msgpack.Decoder) error {
	if err := dec.DecodeMulti(&r.StatusCode, &r.Headers, &r.ttl, &r.ContentLength, &r.body, &r.cachable); err != nil {
		return err
	}
	r.setBodyBytes(r.body)
	return nil
}

// Copy create complete response copy with headers and body
func (r *Response) Copy() (*Response, error) {
	if r == nil {
		return nil, nil
	}

	c := Response{StatusCode: r.StatusCode, ContentLength: r.ContentLength, debug: r.debug, errorValue: r.errorValue}
	c.Headers = r.Headers.Clone()
	c.trans = make([]transforms.Transforms, len(r.trans))
	copy(c.trans, r.trans)
	body, err := r.CopyBody()
	if err != nil {
		return nil, err
	}
	c.setBodyBytes(body)
	return &c, nil

}

// BytesReaderCloser wraps Bytes

type (
	readerAtSeeker interface {
		io.ReaderAt
		io.ReadSeeker
	}
	// bytesReaderAtSeekerNopCloser implements Closer in a way
	// that it preservers readerAtSeeker interface.
	// Helper ioutil.NopCloser narrows the interface to a io.Reader
	// and thus the s3 client spawns extra buffer which is by default 5MB in size.
	bytesReaderAtSeekerNopCloser struct {
		readerAtSeeker
	}
)

func (c bytesReaderAtSeekerNopCloser) Close() error {
	return nil
}

// Stream return io.Reader interface from correct response content
func (r *Response) Stream() io.ReadCloser {
	if r.body != nil {
		return bytesReaderAtSeekerNopCloser{readerAtSeeker: bytes.NewReader(r.body)}
	}

	if r.reader != nil {
		return r.reader
	}

	return nil
}

// BodyTransformer add function that will transform body before send to client
func (r *Response) BodyTransformer(w bodyTransformFnc) {
	r.transformer = w
}

// IsBuffered check if response has access to original buffer
func (r *Response) IsBuffered() bool {
	return r.body != nil
}

// IsImage check if response is image
func (r *Response) IsImage() bool {
	return strings.Contains(r.Headers.Get(HeaderContentType), "image/")
}

func (r *Response) writeDebug() {
	if !r.debug {
		return
	}

	if r.errorValue != nil {

		body := map[string]string{"message": r.errorValue.Error()}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		r.Close()
		r.setBodyBytes(jsonBody)
		r.SetContentType("application/json")
	}
}
