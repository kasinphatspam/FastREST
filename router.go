package fastrest

import (
	"strings"
	"sync"

	"fastrest/context"
)

type Route struct {
	Method     string
	Path       string
	Handlers   []context.Handler
	middleware []context.Middleware
}

type Router struct {
	prefix     string
	routes     *[]*Route
	middleware []context.Middleware
	mu         *sync.RWMutex
}

func newRouter(prefix string) *Router {
	routes := make([]*Route, 0)
	return &Router{
		prefix:     prefix,
		routes:     &routes,
		middleware: make([]context.Middleware, 0),
		mu:         &sync.RWMutex{},
	}
}

func (r *Router) Group(prefix string) *Router {
	return &Router{
		prefix:     r.prefix + prefix,
		routes:     r.routes,
		middleware: append([]context.Middleware{}, r.middleware...),
		mu:         r.mu,
	}
}

func (r *Router) Use(mw ...context.Middleware) {
	r.middleware = append(r.middleware, mw...)
}

func (r *Router) add(method, path string, handlers ...context.Handler) {
	fullPath := r.prefix + path
	route := &Route{
		Method:     method,
		Path:       fullPath,
		Handlers:   handlers,
		middleware: append([]context.Middleware{}, r.middleware...),
	}
	r.mu.Lock()
	*r.routes = append(*r.routes, route)
	r.mu.Unlock()
}

func (r *Router) find(method, path string) (*Route, map[string]string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range *r.routes {
		if route.Method != method {
			continue
		}
		params, ok := matchPath(route.Path, path)
		if ok {
			return route, params
		}
	}
	return nil, nil
}

func matchPath(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = pathParts[i]
		} else if part != pathParts[i] {
			return nil, false
		}
	}
	return params, true
}

func (r *Router) GET(path string, handlers ...context.Handler) {
	r.add("GET", path, handlers...)
}

func (r *Router) POST(path string, handlers ...context.Handler) {
	r.add("POST", path, handlers...)
}

func (r *Router) PUT(path string, handlers ...context.Handler) {
	r.add("PUT", path, handlers...)
}

func (r *Router) PATCH(path string, handlers ...context.Handler) {
	r.add("PATCH", path, handlers...)
}

func (r *Router) DELETE(path string, handlers ...context.Handler) {
	r.add("DELETE", path, handlers...)
}

func (r *Router) HEAD(path string, handlers ...context.Handler) {
	r.add("HEAD", path, handlers...)
}

func (r *Router) OPTIONS(path string, handlers ...context.Handler) {
	r.add("OPTIONS", path, handlers...)
}

func (r *Router) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(*r.routes)
}
