package plugins

import (
	"fmt"
	"net/http"

	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"go.uber.org/zap"
)

// pluginsList list of plugins
var pluginsList = make(map[string]Plugin)

// Plugin interface for Plugins
type Plugin interface {
	preProcess(obj *object.FileObject, req *http.Request)                          // PreProcess is used before start of processing object
	postProcess(obj *object.FileObject, req *http.Request, res *response.Response) // PostProcess is used after end of processing object
	configure(config interface{})
}

// PluginsManager process plugins
type PluginsManager struct {
	list []string
}

// NewPluginsManager create new instance of plugins manager
func NewPluginsManager(plugins map[string]interface{}) PluginsManager {
	pm := PluginsManager{}
	pm.list = make([]string, 0)
	for pName, pConfig := range plugins {
		if _, ok := pluginsList[pName]; !ok {
			panic(fmt.Errorf("unknown plugin %s", pName))
		}

		monitoring.Log().Info("Plugin manager configuring", zap.String("pluginName", pName))
		pluginsList[pName].configure(pConfig)
		pm.list = append(pm.list, pName)
	}
	return pm
}

// PreProcess run PreProcess functions of plugins
func (h PluginsManager) PreProcess(obj *object.FileObject, req *http.Request) {
	for _, hook := range h.list {
		pluginsList[hook].preProcess(obj, req)
	}
}

// PostProcess run PostProcess functions of plugins
func (h PluginsManager) PostProcess(obj *object.FileObject, req *http.Request, res *response.Response) {
	for _, hook := range h.list {
		pluginsList[hook].postProcess(obj, req, res)
	}
}

// RegisterPlugin register plugin
func RegisterPlugin(name string, fnc Plugin) {
	pluginsList[name] = fnc
}
