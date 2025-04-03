package routing

import (
	"net/http"
	"strings"
)

/*
type HTTPHandlerWithContext: A function that takes an http.ResponseWriter, an http.Request, and a Context and returns nothing.
This type is used to define HTTP handlers that can access the context of the request.
Example:

	func MyHandler(w http.ResponseWriter, r *http.Request, c Context) {
	  // Access path parameters
	  id := c.GetPathParam("id")
	  // Access query parameters
	  name := c.GetQueryParam("name")
	  // Handle the request
	}

	func main() {
	  router := routing.NewRouter()
	  router.Handle("GET", "/api/users/:id", MyHandler)
	  router.Serve("8080", routing.ServeOptions{Message: "http://localhost:8080"})
	}
*/
type HTTPHandlerWithContext func(w http.ResponseWriter, r *http.Request, c Context)

type Mux struct {
	routes           map[string]map[string]HTTPHandlerWithContext
	RouterMiddleware []Middleware
	RouteMiddleware  map[string][]Middleware
}

func newMux() *Mux {
	return &Mux{
		routes:           make(map[string]map[string]HTTPHandlerWithContext),
		RouterMiddleware: make([]Middleware, 0),
		RouteMiddleware:  make(map[string][]Middleware),
	}
}

func (m *Mux) handle(method string, path string, handler HTTPHandlerWithContext, middleware ...Middleware) {
	if _, ok := m.routes[method]; !ok {
		m.routes[method] = make(map[string]HTTPHandlerWithContext)
	}

	if _, ok := m.RouteMiddleware[path]; !ok {
		m.RouteMiddleware[path] = make([]Middleware, 0)
	}

	if middleware != nil {
		m.RouteMiddleware[path] = append(m.RouteMiddleware[path], middleware...)
	}
	m.routes[method][path] = handler
}

func (m *Mux) getQueryParams(query string) map[string]string {
	if query == "" {
		return nil
	}

	queryParams := make(map[string]string)
	queryParts := strings.Split(query, "&")

	for _, part := range queryParts {
		parts := strings.Split(part, "=")
		queryParams[parts[0]] = parts[1]
	}
	return queryParams
}

func (m *Mux) matchRoute(r *http.Request, routes map[string]HTTPHandlerWithContext) (HTTPHandlerWithContext, *Context, string) {
	if handler, ok := routes[r.URL.Path]; ok {
		context := newContext()
		return handler, &context, r.URL.Path
	}

	for routePath, handler := range routes {
		if params, ok := m.matchPath(r.URL.Path, routePath); ok {
			context := newContext()
			context.setPathParams(params)
			context.setQueryParams(m.getQueryParams(r.URL.RawQuery))
			return handler, &context, routePath
		}
	}
	return nil, nil, ""
}

func (m *Mux) matchPath(requestPath, routePath string) (map[string]string, bool) {
	routeParts := strings.Split(routePath, "/")
	requestParts := strings.Split(requestPath, "/")

	if len(routeParts) != len(requestParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, part := range routeParts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = requestParts[i]
		} else if part != requestParts[i] {
			return nil, false
		}
	}
	return params, true
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, middleware := range m.RouterMiddleware {
		middleware(w, r)
	}

	routes, ok := m.routes[r.Method]
	if !ok {
		http.NotFound(w, r)
		return
	}

	handler, context, matchedRoute := m.matchRoute(r, routes)
	if handler == nil {
		http.NotFound(w, r)
		return
	}

	if middleware, ok := m.RouteMiddleware[matchedRoute]; ok {
		for _, mw := range middleware {
			mw(w, r)
		}
	}

	handler(w, r, *context)
}
