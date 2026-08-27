package wenovasms

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors returned for local validation failures.
// They are returned before any HTTP request is made and are safe to
// compare with errors.Is.
var (
	// ErrInvalidKey is returned by New when Config.Key is empty.
	ErrInvalidKey = errors.New("wenovasms: api key is required")

	// ErrInvalidPhone is returned when phoneNumber is not a valid Lao
	// mobile number in 20XXXXXXXX format.
	ErrInvalidPhone = errors.New("wenovasms: phone number must be 10 digits starting with 20")

	// ErrInvalidHeader is returned when the SMS header is empty.
	ErrInvalidHeader = errors.New("wenovasms: sms header is required")

	// ErrInvalidMessage is returned when message is empty, longer than
	// 500 characters, or contains a link.
	ErrInvalidMessage = errors.New("wenovasms: message is invalid")
)

// Code is a Wenova business error code (e.g. 30101).
type Code int

// Wenova business error codes. See the Wenova SMS API docs for the full
// list and recommended actions.
const (
	CodeAuthRequired        Code = 30101 // Either token or scriptId is required
	CodeLinkNotAllowed      Code = 30102 // Links are not allowed in SMS messages
	CodePackageInsufficient Code = 30105 // SMS package quota is insufficient
	CodeWalletInsufficient  Code = 30106 // Wallet balance is insufficient
	CodeAccountInactive     Code = 30108 // User account is not active
	CodeWalletNotFound      Code = 30302 // Wallet not found for user
	CodePackageNotFound     Code = 30303 // SMS package not found
	CodeTokenNotFound       Code = 30307 // SMS API token not found
	CodeScriptNotFound      Code = 30308 // SMS scriptId not found
	CodeRateLimited         Code = 30501 // SMS rate limit exceeded
	CodeValidationFailed    Code = 90101 // Request validation failed
	CodeInternalError       Code = 90901 // Internal server error
)

// Error is a Wenova API error. It carries the business Code and the
// HTTP status when the request itself failed.
type Error struct {
	Code   Code   `json:"code"`
	Path   string `json:"path,omitempty"`
	Method string `json:"method,omitempty"`
	// Message is the human-readable error description from Wenova.
	Message string `json:"message,omitempty"`
	// HTTP is the HTTP status code, or 0 when the error did not come
	// from an HTTP response (e.g. a network failure).
	HTTP int `json:"-"`
}

func (e *Error) Error() string {
	if e.HTTP != 0 {
		return fmt.Sprintf("wenovasms: %s (%d) %s", http.StatusText(e.HTTP), e.Code, e.Message)
	}
	return fmt.Sprintf("wenovasms: code %d: %s", e.Code, e.Message)
}

// Is lets errors.Is match an Error against a Code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// GatewayError reports an SMS gateway result that did not succeed
// (resultCode != "20000").
type GatewayError struct {
	ResultCode string
	ResultDesc string
}

func (e *GatewayError) Error() string {
	return fmt.Sprintf("wenovasms: sms gateway rejected the message: %s: %s", e.ResultCode, e.ResultDesc)
}
