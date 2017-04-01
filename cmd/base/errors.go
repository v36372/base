package main

import (
	//"encoding/json"
	"fmt"
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// APIError represents a handler api error
type APIError struct {
	Code    int    `json:"-"`
	Err     error  `json:"-"`
	Message string `json:"message"`
}

// StatusError represent an error with an associated HTTP status code
type StatusError struct {
	Code int
	Err  error
}

// Error allows StatusError to satisfy the error interface
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Status returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

// Status returns our API status code on API.
func (ae APIError) Status() int {
	return ae.Code
}

// Error returns our API error code on API.
func (ae APIError) Error() string {
	return ae.Err.Error()
}

// newAPIError create new API error
func newAPIError(code int, msg string, err error) *APIError {
	if err != nil {
		return &APIError{Code: code, Err: fmt.Errorf(msg+": %s", err), Message: err.Error()}
	}
	return &APIError{Code: code, Err: fmt.Errorf(msg), Message: msg}
}

// newError create new error
func newError(code int, msg string, err error) *StatusError {
	if err != nil {
		return &StatusError{Code: code, Err: fmt.Errorf(msg+": %s", err)}
	}
	return &StatusError{Code: code, Err: fmt.Errorf(msg)}
}

// newSessionSaveError create new session error
func newSessionSaveError(err error) *StatusError {
	return &StatusError{Code: 500, Err: fmt.Errorf("problem saving to cookie store: %s", err)}
}

// newRenderErrMsg create new render error. return string
func newRenderErrMsg(err error) string {
	return fmt.Sprintf("error rendering HTML: %s", err)
}
