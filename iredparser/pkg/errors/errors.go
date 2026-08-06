// Package errors prvides custom application errors
package errors

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ErrTypeAuthentication ErrType = "authentication"
	ErrTypeHTTP           ErrType = "HTTP"
	ErrTypeParsing        ErrType = "parsing"
	ErrTypeCLI            ErrType = "CLI"
)

const (
	// HTTP codes
	ErrCodePostRequestCreation ErrCode = 1001
	ErrCodeGetRequestCreation  ErrCode = 1002
	ErrCodePostRequestFailed   ErrCode = 1003
	ErrCodeGetRequestFailed    ErrCode = 1004
	ErrCodeFailedCaptureCookie ErrCode = 1005

	// Parsig codes
	ErrCodeInvalidMemorySuffix ErrCode = 2001
	ErrCodeCannoFinsCSRFToken  ErrCode = 2002
	ErrCodeChangePassowrd      ErrCode = 2002

	// Authentication codes
	ErrCodeLoginRequired        ErrCode = 3001
	ErrCodeInvalidUsername      ErrCode = 3002
	ErrCodeIncorrectCredentials ErrCode = 3003

	// CLI codes
	ErrCodeCLIUnknown             ErrCode = 4000
	ErrCodeCliInvalidConfig       ErrCode = 4001
	ErrCodeCliInvalidCredentials  ErrCode = 4002
	ErrCodeCliAuthenticationError ErrCode = 4003
)

var (

	// HTTP errors
	ErrPostRequestCreation = New(ErrTypeHTTP, ErrCodePostRequestCreation, errors.New("POST-request creation failed"))
	ErrGetRequestCreation  = New(ErrTypeHTTP, ErrCodeGetRequestCreation, errors.New("GET-request creation failed"))
	ErrPostRequestFailed   = New(ErrTypeHTTP, ErrCodePostRequestFailed, errors.New("POST-request failed"))
	ErrGetRequestFailed    = New(ErrTypeHTTP, ErrCodeGetRequestFailed, errors.New("GET-request failed"))
	ErrFailedCaptureCookie = New(ErrTypeHTTP, ErrCodeFailedCaptureCookie, errors.New("could not capture cookie"))

	// Parsing errors
	ErrInvalidMemorySuffix = New(ErrTypeParsing, ErrCodeInvalidMemorySuffix, errors.New("invalid memory size suffix"))
	ErrCannoFindCSRFToken  = New(ErrTypeParsing, ErrCodeCannoFinsCSRFToken, errors.New("cannot find csrf token"))

	// Authentication errors
	ErrLoginRequired        = New(ErrTypeAuthentication, ErrCodeLoginRequired, errors.New("login required"))
	ErrInvalidUsername      = New(ErrTypeAuthentication, ErrCodeInvalidUsername, errors.New("username must be an valid email address"))
	ErrIncorrectCredentials = New(ErrTypeAuthentication, ErrCodeIncorrectCredentials, errors.New("username or password is incorrect"))

	// CLI errors
	ErrCliUnknown             = New(ErrTypeCLI, ErrCodeCLIUnknown, errors.New("unknown error"))
	ErrCliInvalidConfig       = New(ErrTypeCLI, ErrCodeCliInvalidConfig, errors.New("invalid config"))
	ErrCliInvalidCredentials  = New(ErrTypeCLI, ErrCodeCliInvalidCredentials, errors.New("invalid credentials"))
	ErrCliAuthenticationError = New(ErrTypeCLI, ErrCodeCliAuthenticationError, errors.New("authentication error"))
)

type ErrType string

type ErrCode int

type IRedError struct {
	Type ErrType
	Code ErrCode
	Err  error
}

func (e *IRedError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s:%d] %v", e.Type, e.Code, e.Err)
	}

	return fmt.Sprintf("[%s:%d]", e.Type, e.Code)
}

func New(Type ErrType, Code ErrCode, Err error) *IRedError {
	return &IRedError{
		Type: Type,
		Code: Code,
		Err:  Err,
	}
}

func (e *IRedError) Unwrap() error {
	return e.Err
}

func (e *IRedError) Is(target error) bool {
	var t *IRedError
	if errors.As(target, &t) {
		return e.Type == t.Type
	}

	return false
}

func IsType(err error, errType ErrType) bool {
	var e *IRedError
	if errors.As(err, &e) {
		return e.Type == errType
	}

	return false
}

type IRedMultiError []error

func (m IRedMultiError) Error() string {
	if len(m) == 0 {
		return ""
	}

	msgs := make([]string, len(m))
	for i, err := range m {
		msgs[i] = err.Error()
	}

	return strings.Join(msgs, "; ")
}

func (m IRedMultiError) Unwrap() []error {
	return []error(m)
}
