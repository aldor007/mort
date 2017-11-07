package mort

import (
	"errors"
	"strings"
	"io/ioutil"
	"bytes"

	"mort/config"
	"mort/engine"
	"mort/object"
	"mort/response"
	"mort/storage"
	"mort/transforms"
	"mort/log"
	"strconv"
	"time"
	"net/http"
)

const S3_LOCATION_STR = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><LocationConstraint xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\">EU</LocationConstraint>"

func NewRequestProcessor(max int) RequestProcessor{
	rp := RequestProcessor{}
	rp.Init(max)
	return rp
}

type requestMessage struct {
	responseChan chan *response.Response
	obj *object.FileObject
	request *http.Request
}

type RequestProcessor struct {
	queue chan requestMessage
}

func (r *RequestProcessor) Init(max int)  {
	r.queue = make(chan requestMessage, max)
}

func (r *RequestProcessor) Process(req *http.Request, obj *object.FileObject)  *response.Response{
	msg := requestMessage{}
	msg.request = req
	msg.obj = obj
	msg.responseChan = make(chan *response.Response)

	go r.processChan()
	r.queue <- msg

	select {
	//case <-ctx.Done():
	//	return response.NewBuf(504, "timeout")
	case res := <-msg.responseChan:
		return res
	case <-time.After(time.Second * 60):
		return response.NewBuf(504, []byte("timeout"))
	}
}

func (r *RequestProcessor) processChan()  {
	msg := <- r.queue
	res := r.process(msg.request, msg.obj)
	msg.responseChan <- res
}


func (r *RequestProcessor) process(req *http.Request, obj *object.FileObject) *response.Response {
	switch req.Method {
		case "GET":
			return hanldeGET(req, obj)
		case "PUT":
			return handlePUT(req, obj)

	default:
		return response.NewError(405, errors.New("method not allowed"))
	}
}

func handlePUT(req *http.Request, obj *object.FileObject) *response.Response {
	return storage.Set(obj, req.Header, req.ContentLength, req.Body)
}

func hanldeGET(req *http.Request, obj *object.FileObject) *response.Response {
	if obj.Key == "" {
		return handleS3Get(req, obj)
	}

	var currObj *object.FileObject = obj
	var parentObj *object.FileObject = nil
	var transforms []transforms.Transforms
	var res        *response.Response
	var parentRes  *response.Response

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

	resChan := make(chan *response.Response)
	parentChan := make(chan *response.Response)

	go func(o *object.FileObject) {
		resChan <- storage.Get(o)
	}(obj)

	// get parent from storage
	if parentObj != nil {
		go func(p *object.FileObject) {
			parentChan <- storage.Head(p)
		}(parentObj)
	}

resLoop:
	for {
		select {
		case res = <-resChan:
			if parentObj != nil && (parentRes == nil || parentRes.StatusCode == 0) {
				go func () {
					resChan <- res
				}()

			} else {
				if res.StatusCode == 200 && parentRes.StatusCode == 200 {
					return updateHeaders(res)
				}

				if res.StatusCode == 404 {
					break resLoop
				} else {
					return updateHeaders(res)
				}
			}
		case parentRes = <-parentChan:
			if parentRes.StatusCode == 404 {
				return updateHeaders(parentRes)
			}
		default:

		}
	}

	if obj.HasTransform() && strings.Contains(parentRes.ContentType, "image/") {
		defer res.Close()
		parentRes = updateHeaders(storage.Get(parentObj))

		if parentRes.StatusCode != 200 {
			return updateHeaders(parentRes)
		}

		defer parentRes.Close()

		// revers order of transforms
		for i := 0; i < len(transforms)/2; i++ {
			j := len(transforms) - i - 1
			transforms[i], transforms[j] = transforms[j], transforms[i]
		}

		log.Log().Infow("Performing transforms", "obj.Bucket", obj.Bucket, "obj.Key", obj.Key, "transformsLen", len(transforms))
		return updateHeaders(processImage(obj, parentRes, transforms))
	}

	return updateHeaders(res)
}

func handleS3Get(req *http.Request, obj *object.FileObject) *response.Response {
	query := req.URL.Query()

	if _, ok := query["location"]; ok {
		return response.NewBuf(200, []byte(S3_LOCATION_STR))
	}

	maxKeys := 1000
	delimeter := ""
	prefix := ""
	marker := ""

	if maxKeysQuery, ok := query["max-keys"]; ok {
		maxKeys, _ = strconv.Atoi(maxKeysQuery[0])
	}

	if delimeterQuery, ok := query["delimeter"]; ok {
		delimeter = delimeterQuery[0]
	}

	if prefixQuery, ok := query["prefix"]; ok {
		prefix = prefixQuery[0]
	}

	if markerQuery, ok := query["marker"]; ok {
		marker = markerQuery[0]
	}

	return storage.List(obj, maxKeys, delimeter, prefix, marker)

}

func processImage(obj *object.FileObject, parent *response.Response, transforms []transforms.Transforms) *response.Response {
	engine := engine.NewImageEngine(parent)
	res, err := engine.Process(obj, transforms)
	if err != nil {
		return response.NewError(400, err)
	}

	body, _ := res.CopyBody()
	go func(buf []byte) {
		storage.Set(obj, res.Headers, res.ContentLength, ioutil.NopCloser(bytes.NewReader(buf)))

	}(body)
	return res

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
