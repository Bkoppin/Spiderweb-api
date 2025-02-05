package router

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
	Pattern string
	Handler http.HandlerFunc
	Middleware []Middleware
}

type Router struct {
	Routes []Route
	Middleware []Middleware
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{}
}

// Use applies middleware to all routes
func (r *Router) Use(middleware Middleware) {
	r.Middleware = append(r.Middleware, middleware)
}

func (r *Router) Usemiddleware(next http.HandlerFunc) http.HandlerFunc {
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
func (r *Router) Handle(pattern string, handler http.HandlerFunc) {
	r.Routes = append(r.Routes, Route{Pattern: pattern, Handler: handler})
	http.HandleFunc(pattern, r.Usemiddleware(handler))
}

func (r *Router) Serve(port int) {
	fmt.Printf("Starting server on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}








