// Package errors prvides custom application errors
package errors

import (
	"errors"
	"fmt"
	"slices"
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
	ErrCodeInternalServerError  ErrCode = 1000
	ErrCodePostRequestCreation  ErrCode = 1001
	ErrCodeGetRequestCreation   ErrCode = 1002
	ErrCodePostRequestFailed    ErrCode = 1003
	ErrCodeGetRequestFailed     ErrCode = 1004
	ErrCodeFailedCaptureCookie  ErrCode = 1005
	ErrCodeUnexpectedStatusCode ErrCode = 1006
	ErrCodeStatusCodeNotOK      ErrCode = 1007

	// Parsig codes
	ErrCodeInvalidMemorySuffix ErrCode = 2001
	ErrCodeInvalidQuotaFormat  ErrCode = 2002
	ErrCodeCannoFinsCSRFToken  ErrCode = 2003
	ErrCodeChangePassowrd      ErrCode = 2004
	ErrCodeInvalidMemoryValue  ErrCode = 2005
	ErrCodeEmptyDomain         ErrCode = 2006
	ErrCodeInvalidPageValue    ErrCode = 2007
	ErrCodeEmptyMailAddress    ErrCode = 2008
	ErrCodeDomainParsing       ErrCode = 2009

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
	ErrPostRequestCreation  = New(ErrTypeHTTP, ErrCodePostRequestCreation, errors.New("POST-request creation failed"))
	ErrGetRequestCreation   = New(ErrTypeHTTP, ErrCodeGetRequestCreation, errors.New("GET-request creation failed"))
	ErrPostRequestFailed    = New(ErrTypeHTTP, ErrCodePostRequestFailed, errors.New("POST-request failed"))
	ErrGetRequestFailed     = New(ErrTypeHTTP, ErrCodeGetRequestFailed, errors.New("GET-request failed"))
	ErrFailedCaptureCookie  = New(ErrTypeHTTP, ErrCodeFailedCaptureCookie, errors.New("could not capture cookie"))
	ErrInternalServerError  = New(ErrTypeHTTP, ErrCodeInternalServerError, errors.New("internal server error"))
	ErrUnexpectedStatusCode = New(ErrTypeHTTP, ErrCodeUnexpectedStatusCode, errors.New("unexpected status code"))
	ErrStatusCodeNotOK      = New(ErrTypeHTTP, ErrCodeStatusCodeNotOK, errors.New("status code not OK"))

	// Parsing errors
	ErrInvalidMemorySuffix = New(ErrTypeParsing, ErrCodeInvalidMemorySuffix, errors.New("invalid memory size suffix"))
	ErrCannoFindCSRFToken  = New(ErrTypeParsing, ErrCodeCannoFinsCSRFToken, errors.New("cannot find csrf token"))
	ErrInvalidQuotaFormat  = New(ErrTypeParsing, ErrCodeInvalidQuotaFormat, errors.New("invalid quota format"))
	ErrInvalidMemoryValue  = New(ErrTypeParsing, ErrCodeInvalidMemoryValue, errors.New("invalid memory value"))
	ErrEmptyDomain         = New(ErrTypeParsing, ErrCodeEmptyDomain, errors.New("empty domain"))
	ErrInvalidPageValue    = New(ErrTypeParsing, ErrCodeInvalidPageValue, errors.New("invalid page value"))
	ErrEmptyMailAddress    = New(ErrTypeParsing, ErrCodeEmptyMailAddress, errors.New("empty mail address"))
	ErrDomainParsing       = New(ErrTypeParsing, ErrCodeDomainParsing, errors.New("error while parsing domain"))

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

func (e *IRedError) Wrap(message string) *IRedError {
	if e == nil {
		return nil
	}

	return &IRedError{
		Type: e.Type,
		Code: e.Code,
		Err:  fmt.Errorf("%w: %s", e.Err, message),
	}
}

func (e *IRedError) Wrapf(format string, args ...any) *IRedError {
	if e == nil {
		return nil
	}

	return &IRedError{
		Type: e.Type,
		Code: e.Code,
		Err:  fmt.Errorf("%w: %s", e.Err, fmt.Sprintf(format, args...)),
	}
}

func (e *IRedError) Unwrap() error {
	return e.Err
}

func (e *IRedError) Is(target error) bool {
	var t *IRedError
	if errors.As(target, &t) {
		return e.Type == t.Type && e.Code == t.Code
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

type IRedMultiError struct {
	Errors []error
}

func NewMultiError(errors []error) *IRedMultiError {
	return &IRedMultiError{Errors: errors}
}

func (m *IRedMultiError) Error() string {
	if len(m.Errors) == 0 {
		return ""
	}

	msgs := make([]string, len(m.Errors))
	for i, err := range m.Errors {
		msgs[i] = err.Error()
	}

	return strings.Join(msgs, "; ")
}

func (m *IRedMultiError) Append(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

func (m *IRedMultiError) Unwrap() []error {
	return m.Errors
}

func (m *IRedMultiError) Has(target error) bool {
	if target == nil {
		return len(m.Errors) == 0
	}

	for _, err := range m.Errors {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

func (m *IRedMultiError) HasAny(targets ...error) bool {
	return slices.ContainsFunc(targets, func(target error) bool {
		return m.Has(target)
	})
}
