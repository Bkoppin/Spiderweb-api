package routing

import (
	"fmt"
	"net/http"
)

type Method string

type Middleware func(w http.ResponseWriter, req *http.Request) error

const (
	GET     Method = "GET"
	POST    Method = "POST"
	DELETE  Method = "DELETE"
	PATCH   Method = "PATCH"
	OPTIONS Method = "OPTIONS"
	HEAD    Method = "HEAD"
)

type Route struct {
	Path string
	Handler http.HandlerFunc
	Middleware []Middleware
}

type Router struct {
	Routes []Route
	Middleware []Middleware
}

type ServeOptions struct {
	Message string
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

func (r *Router) useMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, middleware := range r.Middleware {
			err := middleware(w, req)
			if err != nil {
				return
			}
		}
		next(w, req)
}
}

// Handle adds a new route to the router
func (r *Router) Handle(path string, handler http.HandlerFunc) {
	r.Routes = append(r.Routes, Route{Path: path, Handler: handler})
	http.HandleFunc(path, r.useMiddleware(handler))
}

func (r *Router) Serve(port int, options ...ServeOptions) {
	fmt.Printf("Server started on port %d\n", port)
	if len(options) > 0 {
		fmt.Println(options[0].Message)
	}
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	
}








