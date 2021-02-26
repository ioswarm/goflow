package goflow

import (
	"sort"
	"sync"
)

var (
	routeProviderMU sync.RWMutex
	routeProvider   = make(map[string]RouteProvider)
)

func RegisterRouteProvider(name string, provider RouteProvider) {
	routeProviderMU.Lock()
	defer routeProviderMU.Unlock()

	if provider == nil {
		panic("RouteProvider '" + name + "' to register is nil")
	}
	if _, exists := routeProvider[name]; exists {
		panic("RouteProvider " + name + " already registered")
	}
	routeProvider[name] = provider
}

func RouteProviders() []string {
	routeProviderMU.RLock()
	defer routeProviderMU.RUnlock()

	result := make([]string, 0, len(routeProvider))
	for name := range routeProvider {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func newRouteHandler(name string, params ...Parameter) (RouteHandler, error) {
	routeProviderMU.RLock()
	defer routeProviderMU.RUnlock()

	if provider, ok := routeProvider[name]; ok {
		return provider.Create(params...), nil
	}
	return nil, ErrorRouteHandleNotFound
}

type RouteContext interface {
}

type Route struct {
	Path         string
	Method       string
	Handle       interface{}
	Marshaller   Marshal
	Unmarshaller Unmarshal
}

type RouteHandler interface {
	AddRoute(...Route)
}

type RouteProvider interface {
	Create(params ...Parameter) RouteHandler
}
