package routing

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func convertQueryParams(values map[string][]string) map[string]string {
	queryParams := make(map[string]string)
	for key, value := range values {
		if len(value) > 0 {
			queryParams[key] = value[0]
		}
	}
	return queryParams
}

type Context struct {
	PathParams map[string]string
	QueryParams map[string]string
}

type HandlerFunc  func(w http.ResponseWriter, req *http.Request, context Context) http.HandlerFunc

type Middleware func(w http.ResponseWriter, req *http.Request) error

type Route struct {
	Path string
	Handler func(w http.ResponseWriter, req *http.Request, context Context)
	Middleware []Middleware
	Method string
}

type Router struct {
	Routes []Route
	Middleware []Middleware
}

type ServeOptions struct {
	Message string
}

func (c *Context) GetPathParam(key string) string {
	return c.PathParams[key]
}

func (c *Context) GetQueryParam(key string) string {
	return c.QueryParams[key]
}

func newContext(pathParams map[string]string, queryParams map[string]string) Context {
	return Context{
		PathParams: pathParams,
		QueryParams: queryParams,
	}
}

// Creates a new Router object.
func New() *Router {
	return &Router{}
}

/* Use applies middleware to all routes
	 Takes a middleware function as an argument and adds it to the Router's Middleware slice.
*/
func (r *Router) Use(middleware Middleware) {
	r.Middleware = append(r.Middleware, middleware)
}

func getPathParamsValues(path string, routePath string) map[string]string {
	splitPath := strings.Split(path, "/");
	pathParams := make(map[string]string)
	for i, value := range strings.Split(routePath, "/") {
		if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
			pathParams[value[1:len(value)-1]] = splitPath[i]
		}
	}
	return pathParams
}

func (r *Router) useMiddleware(next func(w http.ResponseWriter, req *http.Request, context Context), route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		
		context := newContext(getPathParamsValues(req.URL.Path, route.Path), convertQueryParams(req.URL.Query()))

		for _, middleware := range r.Middleware {
			err := middleware(w, req)
			if err != nil {
				return
			}
		}
		if (req.Method == route.Method) {
		next(w, req, context)
		return
		}

		http.Error(w, "400 Bad Request", http.StatusBadRequest)
	}
}

// Handle adds a new route to the router
func (r *Router) Handle(path string, handler func(w http.ResponseWriter, req *http.Request, context Context), method string) {
	r.Routes = append(r.Routes, Route{Path: path, Handler: handler, Method: method})
	currentRoute := Route{Path: path, Handler: handler, Method: method}
	http.HandleFunc(path, r.useMiddleware(handler, currentRoute))
}

func (r *Router) Serve(port int, options ...ServeOptions) {
	fmt.Printf("Server started on port %d\n", port)
	if len(options) > 0 {
		fmt.Println(options[0].Message)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	
}