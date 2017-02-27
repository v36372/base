package main

import (
	"fmt"
	"log"
	"myapp"
	"net/http"
	"os"
	"time"
)

const (
	// SessionName store key of session
	SessionName = "session-myapp-723322"
	// UserKeyName store key of user
	UserKeyName = "user-myapp-5342086"
	// SessionKeyName store key of session key
	SessionKeyName = "session_key-9532323"
)

// appLogger is an interface for logging.
// Used to introduce a seam into the app, for testing
type appLogger interface {
	Log(str string, v ...interface{})
}

// myappLogger is a wrapper for long.Logger
type myappLogger struct {
	*log.Logger
}

// Log produces a log entry with the current time prepended
func (ml *myappLogger) Log(str string, v ...interface{}) {
	// Prepend current time to the slice of arguments
	v = append(v, 0)
	copy(v[1:], v[0:])
	v[0] = myapp.TimeNow().Format(time.RFC3339)
	ml.Printf("[%s] "+str, v...)
}

// newMiddlewareLogger returns a new middlewareLogger.
func newLogger() *myappLogger {
	return &myappLogger{log.New(os.Stdout, "[myapp] ", 0)}
}

// loggerHanderGenerator prduces a loggingHandler middleware
// loggingHandler middleware logs all request
func (a *App) loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		t1 := myapp.TimeNow()
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

// authMiddleware is the middleware wrapper that detects and provides the user
func (a *App) authMiddleware(db *myapp.DB) func(http http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			var tokenString = req.Header.Get("Authorization")
			// Check token, if it is nil, user must login in system.
			if tokenString == "" {
				a.handleError(w, req, newAPIError(400, "requires authentication ", nil))
				return
			}
			payload, err := DecodePayload(tokenString)
			if err != nil {
				a.handleError(w, req, newAPIError(400, "", err))
				return
			}

			user, err := db.GetUser(payload.Name)
			if err != nil {
				a.logr.Log("panic: %+v", err)
				a.handleError(w, req, newAPIError(404, "user not found ", err))
				return
			}

			session, err := db.GetSessionByUID(user.ID)
			if err != nil {
				a.logr.Log("panic %+v", err)
				a.handleError(w, req, newAPIError(404, "session not found ", err))
			}
			if session.Token == tokenString {
				next.ServeHTTP(w, req)
			} else {
				a.logr.Log("error: token is not valid.")
				a.handleError(w, req, newAPIError(400, "", fmt.Errorf("token is not valid ")))
				return
			}
		}
		return http.HandlerFunc(fn)
	}
}
