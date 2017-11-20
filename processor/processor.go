package processor

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aldor007/mort/config"
	"github.com/aldor007/mort/engine"
	"github.com/aldor007/mort/lock"
	"github.com/aldor007/mort/log"
	"github.com/aldor007/mort/object"
	"github.com/aldor007/mort/response"
	"github.com/aldor007/mort/storage"
	"github.com/aldor007/mort/throttler"
	"github.com/aldor007/mort/transforms"
	"go.uber.org/zap"
)

const s3LocationStr = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><LocationConstraint xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\">EU</LocationConstraint>"

// timeout for collapsed request
var defaultLockTimeout = time.Second * 60

// default request timeout
var defaultProcessTimeout = time.Second * 70

// NewRequestProcessor create instance of request processor
// It main component of mort it handle all of requests
func NewRequestProcessor(queueLen int, l lock.Lock, throttler throttler.Throttler) RequestProcessor {
	rp := RequestProcessor{}
	rp.collapse = l
	rp.throttler = throttler
	rp.queue = make(chan requestMessage, queueLen)

	return rp
}

// RequestProcessor handle incoming requests
type RequestProcessor struct {
	collapse  lock.Lock           // interface used for request collapsing
	throttler throttler.Throttler // interface used for rate limiting creating of new images
	queue     chan requestMessage // request queue
}

type requestMessage struct {
	responseChan chan *response.Response
	obj          *object.FileObject
	request      *http.Request
}

// Process handle incoming request and create response
func (r *RequestProcessor) Process(req *http.Request, obj *object.FileObject) *response.Response {
	msg := requestMessage{}
	msg.request = req
	msg.obj = obj
	msg.responseChan = make(chan *response.Response)
	ctx := req.Context()
	go r.processChan()
	r.queue <- msg

	timer := time.NewTimer(defaultProcessTimeout)
	select {
	case <-ctx.Done():
		log.Log().Warn("Process timeout", zap.String("obj.Key", obj.Key), zap.String("error", "Context.timeout"))
		return response.NewNoContent(499)
	case res := <-msg.responseChan:
		return res
	case <-timer.C:
		log.Log().Warn("Process timeout", zap.String("obj.Key", obj.Key), zap.String("error", "timeout"))
		return response.NewString(504, "timeout")
	}

}

func (r *RequestProcessor) processChan() {
	msg := <-r.queue
	res := r.process(msg.request, msg.obj)
	msg.responseChan <- res
}

func (r *RequestProcessor) process(req *http.Request, obj *object.FileObject) *response.Response {
	switch req.Method {
	case "GET", "HEAD":
		if obj.HasTransform() {
			return r.collapseGET(req, obj)
		}

		return r.handleGET(req, obj)
	case "PUT":
		return handlePUT(req, obj)

	default:
		return response.NewError(405, errors.New("method not allowed"))
	}

}

func handlePUT(req *http.Request, obj *object.FileObject) *response.Response {
	return storage.Set(obj, req.Header, req.ContentLength, req.Body)
}

func (r *RequestProcessor) collapseGET(req *http.Request, obj *object.FileObject) *response.Response {
	ctx := req.Context()
	lockResult, locked := r.collapse.Lock(req.URL.Path)
	if locked {
		log.Log().Info("Lock acquired", zap.String("obj.Key", obj.Key))
		res := r.handleGET(req, obj)
		go r.collapse.NotifyAndRelease(req.URL.Path, res)
		return res
	}

	log.Log().Info("Lock not acquired", zap.String("obj.Key", obj.Key))
	timer := time.NewTimer(defaultLockTimeout)

	for {

		select {
		case <-ctx.Done():
			lockResult.Cancel <- true
			return response.NewNoContent(499)
		case res, ok := <-lockResult.ResponseChan:
			if ok {
				return res
			}

			return r.handleGET(req, obj)
		case <-timer.C:
			lockResult.Cancel <- true
			return response.NewString(504, "timeout")
		default:

		}
	}

}

func (r *RequestProcessor) handleGET(req *http.Request, obj *object.FileObject) *response.Response {
	if obj.Key == "" {
		return handleS3Get(req, obj)
	}

	var currObj *object.FileObject = obj
	var parentObj *object.FileObject
	var transforms []transforms.Transforms
	var res *response.Response
	var parentRes *response.Response
	ctx := req.Context()

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
	if parentObj != nil && obj.CheckParent {
		go func(p *object.FileObject) {
			parentChan <- storage.Head(p)
		}(parentObj)
	}

resLoop:
	for {
		select {
		case <-ctx.Done():
			return response.NewNoContent(499)
		case res = <-resChan:
			if obj.CheckParent && parentObj != nil && (parentRes == nil || parentRes.StatusCode == 0) {
				go func() {
					resChan <- res
				}()

			} else {
				if res.StatusCode == 200 {
					if obj.CheckParent && parentObj != nil && parentRes.StatusCode == 200 {
						return updateHeaders(res)
					}

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

	if parentObj != nil {
		if !obj.CheckParent {
			parentRes = storage.Head(parentObj)
		}

		if obj.HasTransform() && strings.Contains(parentRes.Headers.Get(response.HeaderContentType), "image/") {
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

			log.Log().Info("Performing transforms", zap.String("obj.Bucket", obj.Bucket), zap.String("obj.Key", obj.Key), zap.Int("transformsLen", len(transforms)))
			return updateHeaders(r.processImage(ctx, obj, parentRes, transforms))
		}
	}

	return updateHeaders(res)
}

func handleS3Get(req *http.Request, obj *object.FileObject) *response.Response {
	query := req.URL.Query()

	if _, ok := query["location"]; ok {
		return response.NewString(200, s3LocationStr)
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

func (r *RequestProcessor) processImage(ctx context.Context, obj *object.FileObject, parent *response.Response, transforms []transforms.Transforms) *response.Response {
	taked := r.throttler.Take(ctx)
	if !taked {
		log.Log().Warn("Processor/processImage", zap.String("obj.Key", obj.Key), zap.String("error", "throttled"))
		return response.NewNoContent(503)
	}
	defer r.throttler.Release()

	engine := engine.NewImageEngine(parent)
	res, err := engine.Process(obj, transforms)
	if err != nil {
		return response.NewError(400, err)
	}

	resCpy, err := res.Copy()
	if err == nil {
		go func(objS object.FileObject, resS *response.Response) {
			storage.Set(obj, resS.Headers, resS.ContentLength, resS.Stream())
			resS.Close()
		}(*obj, resCpy)
	} else {
		log.Log().Warn("Processor/processImage", zap.String("obj.Key", obj.Key), zap.Error(err))
	}

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
