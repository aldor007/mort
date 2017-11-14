package mort

import (
	"errors"
	"strings"
	"strconv"
	"time"
	"net/http"
	"context"

	"mort/config"
	"mort/engine"
	"mort/object"
	"mort/response"
	"mort/storage"
	"mort/transforms"
	"mort/log"
	"mort/lock"
	"mort/throttler"
)

const S3_LOCATION_STR = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><LocationConstraint xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\">EU</LocationConstraint>"

func NewRequestProcessor(max int, l lock.Lock) RequestProcessor{
	rp := RequestProcessor{}
	rp.Init(max, l)
	return rp
}

type RequestProcessor struct {
	collapse lock.Lock
	throttler *throttler.Throttler
}

func (r *RequestProcessor) Init(max int, l lock.Lock)  {
	r.collapse = l
	r.throttler = throttler.New(max)
}

func (r *RequestProcessor) Process(req *http.Request, obj *object.FileObject)  *response.Response{
	switch req.Method {
	case "GET", "HEAD":
		if obj.HasTransform() {
			return r.collapseGET(req, obj)
		}

		return r.hanldeGET(req, obj)
	case "PUT":
		return handlePUT(req, obj)

	default:
		return response.NewError(405, errors.New("method not allowed"))
	}

	return response.NewString(503, "ups")
}

func handlePUT(req *http.Request, obj *object.FileObject) *response.Response {
	return storage.Set(obj, req.Header, req.ContentLength, req.Body)
}

func (r *RequestProcessor) collapseGET(req *http.Request, obj *object.FileObject) *response.Response {
	ctx := req.Context()
	resChan, locked := r.collapse.Lock(req.URL.Path)
	if locked {
		log.Log().Infow("Lock acquired", "obj.Key", obj.Key)
		res := r.hanldeGET(req, obj)
		resCpy, _ := res.Copy()
		go r.collapse.NotifyAndRelease(req.URL.Path, resCpy)
		return res
	}

	log.Log().Infow("Lock not acquired", "obj.Key", obj.Key)
	for {

		select {
		case <-ctx.Done():
			return response.NewNoContent(499)
		case res, ok := <- resChan:
			if ok {
				return res
			}

			return r.hanldeGET(req, obj)
		case <-time.After(time.Second * 60):
			return response.NewString(504, "timeout")
		default:

		}
	}

	return response.NewString(500,"ups")

}

func (r *RequestProcessor) hanldeGET(req *http.Request, obj *object.FileObject) *response.Response {
	if obj.Key == "" {
		return handleS3Get(req, obj)
	}

	var currObj *object.FileObject = obj
	var parentObj *object.FileObject = nil
	var transforms []transforms.Transforms
	var res        *response.Response
	var parentRes  *response.Response
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
				go func () {
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

			log.Log().Infow("Performing transforms", "obj.Bucket", obj.Bucket, "obj.Key", obj.Key, "transformsLen", len(transforms))
			return updateHeaders(r.processImage(ctx, obj, parentRes, transforms))
		}
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

func (r *RequestProcessor) processImage(ctx context.Context, obj *object.FileObject, parent *response.Response, transforms []transforms.Transforms) *response.Response {
	taked := r.throttler.Take(ctx)
	if !taked {
		log.Log().Warnw("Processor/processImage", "obj.Key", obj.Key, "error", "throttled")
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
			storage.Set(obj, resS.Headers, resS.ContentLength, resS.Stream)

		}(*obj, resCpy)
	} else {
		log.Log().Warnw("Processor/processImage", "obj.Key", obj.Key, "error", err)
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
