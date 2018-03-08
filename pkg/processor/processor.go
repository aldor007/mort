package processor

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/engine"
	"github.com/aldor007/mort/pkg/helpers"
	"github.com/aldor007/mort/pkg/lock"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/aldor007/mort/pkg/storage"
	"github.com/aldor007/mort/pkg/throttler"
	"github.com/aldor007/mort/pkg/transforms"
	"github.com/karlseguin/ccache"
	"go.uber.org/zap"
)

const s3LocationStr = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><LocationConstraint xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\">EU</LocationConstraint>"

var (
	ErrTimeout       = errors.New("timeout")
	ErrContextCancel = errors.New("context timeout")
	ErrThrottled     = errors.New("throttled")
)

// NewRequestProcessor create instance of request processor
// It main component of mort it handle all of requests
func NewRequestProcessor(serverConfig config.Server, l lock.Lock, throttler throttler.Throttler) RequestProcessor {
	rp := RequestProcessor{}
	rp.collapse = l
	rp.throttler = throttler
	rp.queue = make(chan requestMessage, serverConfig.QueueLen)
	rp.cache = ccache.New(ccache.Configure().MaxSize(serverConfig.CacheSize))
	rp.processTimeout = time.Duration(serverConfig.RequestTimeout) * time.Second
	rp.lockTimeout = time.Duration(serverConfig.RequestTimeout-1) * time.Second
	rp.serverConfig = serverConfig
	return rp
}

// RequestProcessor handle incoming requests
type RequestProcessor struct {
	collapse       lock.Lock           // interface used for request collapsing
	throttler      throttler.Throttler // interface used for rate limiting creating of new images
	queue          chan requestMessage // request queue
	cache          *ccache.Cache       // cache for created image transformations
	processTimeout time.Duration       // request processing timeout
	lockTimeout    time.Duration       // lock timeout for collapsed request it equal processTimeout - 1 s
	serverConfig   config.Server
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

	timer := time.NewTimer(r.processTimeout)
	select {
	case <-ctx.Done():
		monitoring.Log().Warn("Process timeout", zap.String("obj.Key", obj.Key), zap.String("error", "Context.timeout"))
		return r.replyWithError(obj, 499, ErrContextCancel)
	case res := <-msg.responseChan:
		return res
	case <-timer.C:
		monitoring.Log().Warn("Process timeout", zap.String("obj.Key", obj.Key), zap.String("error", "timeout"))
		return r.replyWithError(obj, 504, ErrTimeout)
	}

}

func (r *RequestProcessor) processChan() {
	msg := <-r.queue
	res := r.process(msg.request, msg.obj)
	msg.responseChan <- res
}

func (r *RequestProcessor) replyWithError(obj *object.FileObject, sc int, err error) *response.Response {
	if !obj.HasTransform() || obj.Debug || r.serverConfig.Placeholder == "" {
		return response.NewError(sc, err)
	}

	key := r.serverConfig.Placeholder + strconv.FormatUint(obj.Transforms.Hash().Sum64(), 16) + strconv.FormatInt(int64(sc), 10)
	if cacheRes := r.fetchResponseFromCache(key); cacheRes != nil {
		return cacheRes
	}

	buf, err := helpers.FetchObject(r.serverConfig.Placeholder)
	if err != nil {
		return response.NewError(sc, err)
	}

	parent := response.NewBuf(sc, buf)
	transformsTab := []transforms.Transforms{obj.Transforms}

	eng := engine.NewImageEngine(parent)
	res, err := eng.Process(obj, transformsTab)
	res.StatusCode = sc
	resCpy, errCpy := res.Copy()
	if errCpy != nil {
		r.cache.Set(key, resCpy, time.Minute*10)
	}

	return res
}

func (r *RequestProcessor) process(req *http.Request, obj *object.FileObject) *response.Response {
	switch req.Method {
	case "GET", "HEAD":
		if obj.Key == "" {
			return handleS3Get(req, obj)
		}

		if obj.HasTransform() {
			return updateHeaders(req, r.collapseGET(req, obj))
		}

		return updateHeaders(req, r.handleGET(req, obj))
	case "PUT":
		return handlePUT(req, obj)
	case "DELETE":
		return storage.Delete(obj)

	default:
		return response.NewError(405, errors.New("method not allowed"))
	}

}

func handlePUT(req *http.Request, obj *object.FileObject) *response.Response {
	return storage.Set(obj, req.Header, req.ContentLength, req.Body)
}

func (r *RequestProcessor) collapseGET(req *http.Request, obj *object.FileObject) *response.Response {
	ctx := req.Context()
	lockResult, locked := r.collapse.Lock(obj.Key)
	if locked {
		monitoring.Log().Info("Lock acquired", zap.String("obj.Key", obj.Key))
		res := r.handleGET(req, obj)
		r.collapse.NotifyAndRelease(obj.Key, res)
		return res
	}

	monitoring.Report().Inc("collapsed_count")
	monitoring.Log().Info("Lock not acquired", zap.String("obj.Key", obj.Key))
	timer := time.NewTimer(r.lockTimeout)

	for {

		select {
		case <-ctx.Done():
			lockResult.Cancel <- true
			return r.replyWithError(obj, 499, ErrContextCancel)
		case res, ok := <-lockResult.ResponseChan:
			if ok {
				return res
			}

			return r.handleGET(req, obj)
		case <-timer.C:
			lockResult.Cancel <- true
			return r.replyWithError(obj, 504, ErrTimeout)
		default:
			if cacheRes := r.fetchResponseFromCache(obj.Key); cacheRes != nil {
				lockResult.Cancel <- true
				return cacheRes
			}
		}
	}

}

func (r *RequestProcessor) fetchResponseFromCache(key string) *response.Response {
	cacheValue := r.cache.Get(key)
	if cacheValue != nil {
		if cacheValue.Expired() == false {
			monitoring.Log().Info("Handle Get cache", zap.String("cache", "hit"), zap.String("obj.Key", key))
			monitoring.Report().Inc("cache_ratio;status:hit")
			res := cacheValue.Value().(*response.Response)
			resCp, err := res.Copy()
			if err == nil {
				return resCp
			}

		} else {
			monitoring.Log().Info("Handle Get cache", zap.String("cache", "expired"), zap.String("obj.Key", key))
			monitoring.Report().Inc("cache_ratio;status:expired")
			res := cacheValue.Value().(*response.Response)
			res.Close()
			r.cache.Delete(key)
		}
	}

	return nil

}

func (r *RequestProcessor) handleGET(req *http.Request, obj *object.FileObject) *response.Response {
	if cacheRes := r.fetchResponseFromCache(obj.Key); cacheRes != nil {
		return cacheRes
	}

	var currObj *object.FileObject = obj
	var parentObj *object.FileObject
	var transformsTab []transforms.Transforms
	var res *response.Response
	var parentRes *response.Response
	ctx := req.Context()

	// search for last parent
	for currObj.HasParent() {
		if currObj.HasTransform() {
			transformsTab = append(transformsTab, currObj.Transforms)
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
			return r.replyWithError(obj, 499, ErrContextCancel)
		case res = <-resChan:
			if obj.CheckParent && parentObj != nil && (parentRes == nil || parentRes.StatusCode == 0) {
				go func() {
					resChan <- res
				}()

			} else {
				if res.StatusCode == 200 {
					if obj.CheckParent && parentObj != nil && parentRes.StatusCode == 200 {
						return res
					}

					return res
				}

				if res.StatusCode == 404 {
					break resLoop
				} else {
					return res
				}
			}
		case parentRes = <-parentChan:
			if parentRes.StatusCode == 404 {
				return parentRes
			}
		default:

		}
	}

	if parentObj != nil {
		if !obj.CheckParent {
			parentRes = storage.Head(parentObj)
		}

		if parentRes.HasError() {
			return r.replyWithError(obj, parentRes.StatusCode, parentRes.Error())
		} else if parentRes.StatusCode == 404 {
			return parentRes
		}

		if obj.HasTransform() && parentRes.StatusCode == 200 && strings.Contains(parentRes.Headers.Get(response.HeaderContentType), "image/") {
			defer res.Close()
			parentRes.Close()
			parentRes = storage.Get(parentObj)

			defer parentRes.Close()

			transLen := len(transformsTab)
			if transLen > 1 {
				// revers order of transforms
				for i := 0; i < len(transformsTab)/2; i++ {
					j := len(transformsTab) - i - 1
					transformsTab[i], transformsTab[j] = transformsTab[j], transformsTab[i]
				}

			}

			monitoring.Log().Info("Performing transforms", zap.String("obj.Bucket", obj.Bucket), zap.String("obj.Key", obj.Key), zap.Int("transformsLen", len(transformsTab)))
			return r.processImage(ctx, obj, parentRes, transformsTab)
		} else if obj.HasTransform() {
			parentRes.Close()
			monitoring.Log().Warn("Not performing transforms", zap.String("obj.Bucket", obj.Bucket), zap.String("obj.Key", obj.Key),
				zap.String("parent.Key", parentObj.Key), zap.Int("parent.sc", parentRes.StatusCode), zap.String("parent.ContentType", parentRes.Headers.Get(response.HeaderContentType)), zap.Error(parentRes.Error()))
		}

	}

	return res
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
		monitoring.Log().Warn("Processor/processImage", zap.String("obj.Key", obj.Key), zap.String("error", "throttled"))
		monitoring.Report().Inc("throttled_count")
		return r.replyWithError(obj, 503, ErrThrottled)
	}
	defer r.throttler.Release()

	eng := engine.NewImageEngine(parent)
	res, err := eng.Process(obj, transforms)
	if err != nil {
		return response.NewError(400, err)
	}

	resCpy, err := res.Copy()
	if err == nil {
		r.cache.Set(obj.Key, resCpy, time.Minute*2)
		go func(objS object.FileObject, resS *response.Response) {
			storage.Set(&objS, resS.Headers, resS.ContentLength, resS.Stream())
			//r.cache.Delete(objS.Key)
			resS.Close()
		}(*obj, resCpy)
	} else {
		monitoring.Log().Warn("Processor/processImage", zap.String("obj.Key", obj.Key), zap.Error(err))
	}

	return res

}

func updateHeaders(req *http.Request, res *response.Response) *response.Response {
	ctx := req.Context()

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

	if ctx.Value("auth") != nil {
		res.Set("Cache-Control", "no-cache")
	}

	return res
}
