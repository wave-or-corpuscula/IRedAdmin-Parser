// Package errors prvides custom application errors
package errors

import "errors"

var (
	ErrAuthenticationFailed = errors.New("authenticatoin failed")
	ErrPostRequestCreation  = errors.New("POST-request creation failed")
	ErrGetRequestCreation   = errors.New("GET-request creation failed")
	ErrPostRequestFailed    = errors.New("POST-request failed")
	ErrGetRequestFailed     = errors.New("GET-request failed")
	ErrFailedCaptureCookie  = errors.New("could not capture cookie")
	ErrInvalidResponseData  = errors.New("invalid response data")
	ErrInvalidMemorySuffix  = errors.New("invalid memory size suffix")
	ErrInvalidCredentials   = errors.New("invalid credentials")
)
