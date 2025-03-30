package routing

import (
	"fmt"
	stdlog "log"
	"net/http"
	"os"

	admissioncontrol "github.com/elithrar/admission-control"
	"github.com/go-kit/log"
)

type Middleware func(http.ResponseWriter, *http.Request)

type Context struct {
	PathParams map[string]string
	QueryParams map[string]string
}

type ServeOptions struct {
	Message string
}
type Router struct {
	middleware []Middleware
	mux *Mux
}

type Route struct {
	Method string
	Path string
	Handler HTTPHandlerWithContext
	Middleware []Middleware
}

func newContext() Context {
	return Context{
		PathParams: make(map[string]string),
		QueryParams: make(map[string]string),
	}
}

func (c *Context) setPathParams(paramMap map[string]string) {
	c.PathParams = paramMap
}

func (c *Context) setQueryParams(paramMap map[string]string) {
	c.QueryParams = paramMap
}

func (c Context) GetPathParam(key string) string {
	return c.PathParams[key]
}

func (c Context) GetQueryParam(key string) string {
	return c.QueryParams[key]
}

func NewRouter() *Router {
	

	return &Router{
		middleware: make([]Middleware, 0),
		mux: newMux(),
	}
}

func (r *Router) Use(m Middleware) {
	r.middleware = append(r.middleware, m)
	r.mux.RouterMiddleware = r.middleware
}

func (r *Router) Handle(method string, path string, handler HTTPHandlerWithContext, middleware ...Middleware) *Route {
	route := Route{
		Method: method,
		Path: path,
		Handler: handler,
		Middleware: middleware,
	}
	r.mux.handle(method, path, handler, middleware...)
	
	return &route
}

func (r *Router) Serve(port string, options ServeOptions) error {
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
	return nil
}
