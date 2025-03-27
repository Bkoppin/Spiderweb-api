package routing

import (
	"fmt"
	"net/http"
	"strings"
)


type HTTPHandlerWithContext func(w http.ResponseWriter, r *http.Request, c Context)

type Mux struct {
	routes map[string]map[string]HTTPHandlerWithContext
	RouterMiddleware []Middleware
	RouteMiddleware map[string][]Middleware
}


func newMux() *Mux {
	return &Mux{
		routes: make(map[string]map[string]HTTPHandlerWithContext),
		RouterMiddleware: make([]Middleware, 0),
		RouteMiddleware: make(map[string][]Middleware),
	}
}


func (m *Mux) handle(method string, path string, handler HTTPHandlerWithContext, middleware ...Middleware) {
	if _, ok := m.routes[method]; !ok {
		m.routes[method] = make(map[string]HTTPHandlerWithContext)
	}

	if _, ok := m.RouteMiddleware[path]; !ok {
		m.RouteMiddleware[path] = make([]Middleware, 0)
	}

	if (middleware != nil) {
	m.RouteMiddleware[path] = append(m.RouteMiddleware[path], middleware...)
	}
	m.routes[method][path] = handler
}

func (m *Mux) getPathParams(path string, route string) map[string]string {
	pathParams := make(map[string]string)
	pathParts := strings.Split(path, "/")
	routeParts := strings.Split(route, "/")
	if len(pathParts) == len(routeParts) {
	for i, part := range routeParts {
		if strings.HasPrefix(part, ":") {
			pathParams[part[1:]] = pathParts[i]
		}
	}
	}
	
	return pathParams
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



func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, middleware := range m.RouterMiddleware {
		middleware(w, r)
	}
	if routes, ok := m.routes[r.Method]; ok {
		if handler, ok := routes[r.URL.Path]; ok {
			context := newContext()
			for _, middleware := range m.RouteMiddleware[r.URL.Path] {
				middleware(w, r)
			}
			handler(w, r, context)
			return
		}
	for routePath, handler := range routes {
		routeParts := strings.Split(routePath, "/")
		requestParts := strings.Split(r.URL.Path, "/")
		if len(routeParts) == len(requestParts) {
			match := true
			params := make(map[string]string)
			for i, part := range routeParts {
							if strings.HasPrefix(part, ":") {
											params[part[1:]] = requestParts[i]
							} else if part != requestParts[i] {
											match = false
											break
							}
			}
			if match {
				fmt.Println("Matched route", routePath)
				pathParams := m.getPathParams(r.URL.Path, routePath)
				queryParams := m.getQueryParams(r.URL.RawQuery)
				
				context := newContext()
				context.setPathParams(pathParams)
				context.setQueryParams(queryParams)

				for _, middleware := range m.RouteMiddleware[routePath] {
					middleware(w, r)
				}
				handler(w, r, context)
				return
			}
		}
	}
}
	http.NotFound(w, r)
}

