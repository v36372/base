package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"base"
)

// appLogger is an interface for logging.
// Used to introduce a seam into the app, for testing
type appLogger interface {
	Log(str string, v ...interface{})
}

// baseLogger is a wrapper for long.Logger
type baseLogger struct {
	*log.Logger
}

// Log produces a log entry with the current time prepended
func (ml *baseLogger) Log(str string, v ...interface{}) {
	// Prepend current time to the slice of arguments
	v = append(v, 0)
	copy(v[1:], v[0:])
	v[0] = base.TimeNow().Format(time.RFC3339)
	ml.Printf("[%s] "+str, v...)
}

// newMiddlewareLogger returns a new middlewareLogger.
func newLogger() *baseLogger {
	return &baseLogger{log.New(os.Stdout, "[base] ", 0)}
}

// loggerHanderGenerator prduces a loggingHandler middleware
// loggingHandler middleware logs all request
func (a *App) loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		t1 := base.TimeNow()
		a.logr.Log("Started %s %s", req.Method, req.URL.Path)

		next.ServeHTTP(w, req)

		rw := w.(ResponseWriter)
		a.logr.Log("Completed %v %s in %v", rw.Status(), http.StatusText(rw.Status()), time.Since(t1))
	}
	return http.HandlerFunc(fn)
}

// recoverHandlerGenerator products a recoverHandler middleware
// recoverHander is an middleware that captures and recovers from panics
func (a *App) recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.logr.Log("Panic: %+v", err)
				//
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
