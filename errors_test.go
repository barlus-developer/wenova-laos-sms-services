package wenovasms

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorFormat(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "with http status",
			err:  &Error{Code: CodeRateLimited, Message: "too many", HTTP: 429},
			want: "wenovasms: Too Many Requests (30501) too many",
		},
		{
			name: "without http status",
			err:  &Error{Code: CodeInternalError, Message: "boom"},
			want: "wenovasms: code 90901: boom",
		},
		{
			name: "empty message",
			err:  &Error{Code: CodeAuthRequired, HTTP: 400},
			want: "wenovasms: Bad Request (30101) ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	base := &Error{Code: CodeAuthRequired, Message: "x"}
	same := &Error{Code: CodeAuthRequired, Message: "y"}
	different := &Error{Code: CodeRateLimited, Message: "z"}

	if !errors.Is(base, same) {
		t.Error("errors with the same code should match")
	}
	if errors.Is(base, different) {
		t.Error("errors with different codes should not match")
	}
	if errors.Is(base, ErrInvalidKey) {
		t.Error("an *Error must not match a sentinel validation error")
	}
}

func TestGatewayErrorFormat(t *testing.T) {
	err := &GatewayError{ResultCode: "30001", ResultDesc: "invalid message"}
	if !strings.Contains(err.Error(), "30001") {
		t.Errorf("Error() should include the result code, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "invalid message") {
		t.Errorf("Error() should include the result description, got %q", err.Error())
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{ErrInvalidKey, "api key"},
		{ErrInvalidPhone, "phone"},
		{ErrInvalidHeader, "header"},
		{ErrInvalidMessage, "message"},
	}

	for _, tt := range tests {
		if !strings.Contains(tt.err.Error(), tt.want) {
			t.Errorf("%v should mention %q", tt.err, tt.want)
		}
	}
}
