package main

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

// Router is a wrapper around httprouter
// Here we reimplement the type requests, but with a
// wrapHandler
type Router struct {
	*httprouter.Router
}

// NewRouter return a new router
func NewRouter() *Router {
	return &Router{httprouter.New()}
}

// Params ...
const Params = "params"

// wrapHandler turns a normal http.Handler into a httprouter compatible
// handler. We use gorilla/context to save params instead.
// This incurs a small performance hit, but it allows us to conform to the
// http.Handler interface.
func wrapHandler(next http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		context.Set(req, Params, ps)
		// Use our own ResponseWriter wrapper in order to capture response data.
		next.ServeHTTP(NewResponseWriter(w), req)
	}
}

// Get presenter for GET
func (r *Router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

// Post presenter for POST
func (r *Router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

// Put presenter for PUT
func (r *Router) Put(path string, handler http.Handler) {
	r.PUT(path, wrapHandler(handler))
}

// Patch presenter for PATCH
func (r *Router) Patch(path string, handler http.Handler) {
	r.PATCH(path, wrapHandler(handler))
}

// Delete presenter for DELETE
func (r *Router) Delete(path string, handler http.Handler) {
	r.DELETE(path, wrapHandler(handler))
}

// Head presenter for HEAD
func (r *Router) Head(path string, handler http.Handler) {
	r.HEAD(path, wrapHandler(handler))
}

// Options presenter for OPTIONS
func (r *Router) Options(path string, handler http.Handler) {
	r.OPTIONS(path, wrapHandler(handler))
}
