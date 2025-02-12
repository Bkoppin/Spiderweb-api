/*
Routing package is a simple http router with similar syntax to express.js

	func main() {
		router := routing.New()
		router.Use(middleware.Cors)
		router.Use(middleware.ContentTypeJSON)
		router.Handle("/api/users", controller.CreateUser, "POST")
		router.Serve(8080, routing.ServeOptions{Message: "http://localhost:8080",})
	}
*/
package routing

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)
type Context struct {
	PathParams map[string]string
	QueryParams map[string]string
}

type HandlerFunc  func(w http.ResponseWriter, req *http.Request, context Context) http.HandlerFunc

type Middleware func(w http.ResponseWriter, req *http.Request) error

type Route struct {
	Path string
	Handler func(w http.ResponseWriter, req *http.Request, context Context)
	Middleware Middleware
	Method string
}

type Router struct {
	Routes []Route
	Middleware []Middleware
}

type ServeOptions struct {
	Message string
}

func convertQueryParams(values map[string][]string) map[string]string {
	queryParams := make(map[string]string)
	for key, value := range values {
		if len(value) > 0 {
			queryParams[key] = value[0]
		}
	}
	return queryParams
}

/*
	GetPathParam returns the value of a path parameter

	@param key string - the key of the path parameter

	@return string - the value of the path parameter

	func GetUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
	 id := context.GetPathParam("id")
	 fmt.Println(id)
	}
*/
func (c *Context) GetPathParam(key string) string {
	return c.PathParams[key]
}

/*
	GetQueryParam returns the value of a query parameter

	@param key string - the key of the query parameter

	@return string - the value of the query parameter

	func GetUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
	 id := context.GetQueryParam("id")
	 fmt.Println(id)
	}
*/
func (c *Context) GetQueryParam(key string) string {
	return c.QueryParams[key]
}

func newContext(pathParams map[string]string, queryParams map[string]string) Context {
	return Context{
		PathParams: pathParams,
		QueryParams: queryParams,
	}
}

/* 
	Create a new Router

	type Router struct {
		Routes []Route
		Middleware []Middleware
	}

	@return *Router - the new Router

	func main() {
		router := routing.New()
		router.Use(middleware.Cors)
		router.Use(middleware.ContentTypeJSON)
		router.Handle("/api/users", controller.CreateUser, "POST")
	}

*/
func New() *Router {
	return &Router{}
}

/*Use applies middleware to all routes on the current Router

	type Middleware func(w http.ResponseWriter, req *http.Request) error
	func main() {
		router := routing.New()
		router.Use(middleware.Cors)
		router.Use(middleware.ContentTypeJSON)
		router.Handle("/api/users", controller.CreateUser, "POST")
}
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
		if req.Method != route.Method {
			http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if route.Middleware == nil {
			next(w, req, context)
			return
		}

		err := route.Middleware(w, req)
		if err != nil {
			return
		}


		next(w, req, context)

	}
}

/*
	Handle creates a new route on the Router
	
	@param path string - the path of the route

	@param handler func(w http.ResponseWriter, req *http.Request, context Context) - the handler for the route

	@param method string - the method of the route

	@param routeMiddleware ...Middleware - the middleware for the route	(optional)

	@return Route - the new route


	func main() {
		router := routing.New()
		router.Handle("/api/users", controller.CreateUser, "POST", middleware.WithAuth)
		router.Handle("/api/users/{id}", controller.GetUser, "GET")
	}
*/
func (r *Router) Handle(path string, handler func(w http.ResponseWriter, req *http.Request, context Context), method string, routeMiddleware ...Middleware) Route {
	var middleware Middleware
	if len(routeMiddleware) > 0 {
		middleware = routeMiddleware[0]
	}
	r.Routes = append(r.Routes, Route{Path: path, Handler: handler, Method: method, Middleware: middleware})
	route := r.Routes[len(r.Routes)-1]
	http.HandleFunc(path, r.useMiddleware(handler, route))
	return route
}

/*
	Serve starts the server on the specified port

	@param port int - the port to start the server on

	@param options ...ServeOptions - the options for the server

	@return void

	func main() {
		router := routing.New()
		router.Serve(8080, routing.ServeOptions{Message: "http://localhost:8080",})
	}
		
*/
func (r *Router) Serve(port int, options ...ServeOptions) {
	fmt.Printf("Server started on port %d\n", port)
	if len(options) > 0 {
		fmt.Println(options[0].Message)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	
}

