// Package routing is a lightweight HTTP router that allows for middleware chaining and context management.
//
// Included public types and functions:
//
//   - @package routing
//
//   - @type Middleware - A function that takes an http.ResponseWriter and an http.Request and returns nothing.
//
//   - @type Context - A struct that holds path and query parameters.
//
//   - @type ServeOptions - A struct that holds options for serving the router.
//
//   - @type Router - A struct that holds middleware and a Mux instance.
//
//   - @type Route - A struct that holds HTTP method, path, handler, and middleware for a specific route.
//
//   - @type HTTPHandlerWithContext - A function that takes an http.ResponseWriter, an http.Request, and a Context and returns nothing.
//
//   - @func Use - Adds a middleware to the Router's middleware chain.
//
//   - @func Handle - Registers a route with the specified method, path, handler, and middleware.
//
//   - @func Serve - Starts the HTTP server on the specified port with the provided options.
package routing

import (
	"fmt"
	stdlog "log"
	"net/http"
	"os"

	admissioncontrol "github.com/elithrar/admission-control"
	"github.com/go-kit/log"
)

/*
type Middleware: A function that takes an http.ResponseWriter and an http.Request and returns nothing.

This type is used to define middleware functions that can be applied to HTTP routes.
*/
type Middleware func(http.ResponseWriter, *http.Request)

/*
type Context: A struct that holds path and query parameters.

This struct is used to manage the context of an HTTP request, including path parameters and query parameters.

  - @property PathParams: A map of path parameters, where the key is the parameter name and the value is the parameter value.
  - @property QueryParams: A map of query parameters, where the key is the parameter name and the value is the parameter value.
  - @method @private setPathParams: Sets the path parameters for the context.
  - @method @private setQueryParams: Sets the query parameters for the context.
  - @method GetPathParam: Returns the value of a path parameter by its key.
  - @method GetQueryParam: Returns the value of a query parameter by its key.
  - @constructor @private newContext: Creates a new Context instance with empty path and query parameters.
*/
type Context struct {
	PathParams  map[string]string
	QueryParams map[string]string
}

/*
type ServeOptions: A struct that holds options for serving the router.
This struct is used to configure the HTTP server when it is started.
  - @property Message: A message to be displayed when the server starts.
*/
type ServeOptions struct {
	Message string
	Logging bool
}
/*
type Router: A struct that holds middleware and a Mux instance.
This struct is used to manage the routing of HTTP requests and apply middleware to routes.
	- @property middleware: A slice of Middleware functions to be applied to the router.
	- @property mux: A Mux instance that handles the actual routing of HTTP requests.
*/
type Router struct {
	middleware []Middleware
	mux        *Mux
}

/*
type Route: A struct that holds HTTP method, path, handler, and middleware for a specific route.
This struct is used to define a route in the router.
  - @property Method: The HTTP method for the route (e.g., GET, POST).
  - @property Path: The path for the route (e.g., /api/v1/resource).
  - @property Handler: The handler function for the route, which takes an http.ResponseWriter, an http.Request, and a Context.
  - @property Middleware: A slice of middleware functions to be applied to the route.
*/
type Route struct {
	Method     string
	Path       string
	Handler    HTTPHandlerWithContext
	Middleware []Middleware
}

/*
func newContext: Creates a new Context instance with empty path and query parameters.
This function initializes a new Context struct with empty maps for path and query parameters.
  - @return: A new Context instance.
*/
func newContext() Context {
	return Context{
		PathParams:  make(map[string]string),
		QueryParams: make(map[string]string),
	}
}

/*
func (c *Context) setPathParams: Sets the path parameters for the context.
This method updates the PathParams map in the Context struct with the provided parameter map.
  - @param paramMap: A map of path parameters, where the key is the parameter name and the value is the parameter value.
*/
func (c *Context) setPathParams(paramMap map[string]string) {
	c.PathParams = paramMap
}

/*
func (c *Context) setQueryParams: Sets the query parameters for the context.
This method updates the QueryParams map in the Context struct with the provided parameter map.
  - @param paramMap: A map of query parameters, where the key is the parameter name and the value is the parameter value.
*/
func (c *Context) setQueryParams(paramMap map[string]string) {
	c.QueryParams = paramMap
}

/*
func (c Context) GetPathParam: Returns the value of a path parameter by its key.
This method retrieves the value of a path parameter from the PathParams map in the Context struct.
  - @param key: The key of the path parameter to retrieve.
  - @return: The value of the specified path parameter.
Example usage:
	func myHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
		id := ctx.GetPathParam("id")
		// Use the id parameter for further processing
	}

*/
func (c Context) GetPathParam(key string) string {
	return c.PathParams[key]
}

/*
func (c Context) GetQueryParam: Returns the value of a query parameter by its key.
This method retrieves the value of a query parameter from the QueryParams map in the Context struct.
  - @param key: The key of the query parameter to retrieve.
  - @return: The value of the specified query parameter.

Example usage:
  func myHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
		id := ctx.GetQueryParam("sort")
		// Use the id parameter for further processing
	}
*/
func (c Context) GetQueryParam(key string) string {
	return c.QueryParams[key]
}

/*
func NewRouter: Creates a new Router instance with an empty middleware chain and a new Mux instance.
This function initializes a Router struct with an empty slice of middleware and a new Mux instance.
  - @return: A new Router instance.

Example usage:
	router := NewRouter()
	router.Use(myMiddleware)
	router.Handle("GET", "/api/v1/resource", myHandler)
	router.Serve("8080", ServeOptions{Message: "Server started on port 8080"})
*/
func NewRouter() *Router {
	return &Router{
		middleware: make([]Middleware, 0),
		mux:        newMux(),
	}
}

// Use adds a middleware to the Router's middleware chain and updates the
// Router's internal mux with the new middleware list.
//
// Parameters:
//   - m: The middleware to be added to the Router's middleware chain.
//   - This method appends the provided middleware to the Router's middleware slice and updates the mux.RouterMiddleware with the new middleware list.
//   - This allows the middleware to be applied to all routes handled by the Router.
// Example usage:
//   router := NewRouter()
//   router.Use(myMiddleware)
//   router.Handle("GET", "/api/v1/resource", myHandler)
func (r *Router) Use(m Middleware) {
	r.middleware = append(r.middleware, m)
	r.mux.RouterMiddleware = r.middleware
}

/*
func (r *Router) Handle: Registers a route with the specified method, path, handler, and middleware.
This method adds a new route to the Router's internal mux and returns a Route instance.
  - @param method: The HTTP method for the route (e.g., GET, POST).
  - @param path: The path for the route (e.g., /api/v1/resource).
  - @param handler: The handler function for the route, which takes an http.ResponseWriter, an http.Request, and a Context.
  - @param middleware: A variadic list of middleware functions to be applied to the route.
  - @return: A Route instance representing the registered route.

Example usage:
	router := NewRouter()
	router.Handle("GET", "/api/v1/resource", myHandler, myMiddleware1, myMiddleware2)
*/
func (r *Router) Handle(method string, path string, handler HTTPHandlerWithContext, middleware ...Middleware) *Route {
	route := Route{
		Method:     method,
		Path:       path,
		Handler:    handler,
		Middleware: middleware,
	}
	r.mux.handle(method, path, handler, middleware...)

	return &route
}

/*
func (r *Router) Serve: Starts the HTTP server on the specified port with the provided options.
This method initializes the server with the specified port and options, and starts listening for incoming HTTP requests.
  - @param port: The port on which the server will listen for incoming requests.
  - @param options: A ServeOptions instance containing options for serving the router.
  - @return: An error if the server fails to start.

Example usage:
	router := NewRouter()
	router.Serve("8080", ServeOptions{Message: "Server started on port 8080"})

*/
func (r *Router) Serve(port string, options ServeOptions) error {
	if options.Logging {
		var logger log.Logger

		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		stdlog.SetOutput(log.NewStdlibAdapter(logger))

		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "loc", log.DefaultCaller)

		loggingMiddleware := admissioncontrol.LoggingMiddleware(logger)

		fmt.Println("Server started on port", port)
		fmt.Println("Message:", options.Message)
		if err := http.ListenAndServe(":"+port, loggingMiddleware(r.mux)); err != nil {
			logger.Log("status", "fatal", "err", err)
			os.Exit(1)
		}
	}
	fmt.Println("Server started on port", port)
	fmt.Println("Message:", options.Message)
	http.ListenAndServe(":"+port, r.mux)
	return nil
}
