package processor

import (
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"net/http"
)

// hooksList list of plugins
var hooksList = make(map[string]Hook)

// Hook interface for Plugins
type Hook interface {
	preProcess(obj *object.FileObject, req *http.Request)                          // preProcess is used before start of processing object
	postProcess(obj *object.FileObject, req *http.Request, res *response.Response) // postProcess is used after end of processing object
}

// HooksProcessor process plugins
type HooksProcessor struct {
	list []string
}

// NewHooksProcessor create new instance of plugins manager
func NewHooksProcessor(plugins []string) HooksProcessor {
	return HooksProcessor{plugins}
}

// preProcess run preProcess functions of plugins
func (h HooksProcessor) preProcess(obj *object.FileObject, req *http.Request) {
	for _, hook := range h.list {
		hooksList[hook].preProcess(obj, req)
	}
}

// postProcess run postProcess functions of plugins
func (h HooksProcessor) postProcess(obj *object.FileObject, req *http.Request, res *response.Response) {
	for _, hook := range h.list {
		hooksList[hook].postProcess(obj, req, res)
	}
}

// RegisterHook register plugin
func RegisterHook(name string, fnc Hook) {
	hooksList[name] = fnc
}
