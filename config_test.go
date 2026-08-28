package wenovasms

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{name: "empty key", cfg: Config{}, wantErr: ErrInvalidKey},
		{name: "whitespace key", cfg: Config{Key: "   "}, wantErr: ErrInvalidKey},
		{name: "valid key", cfg: Config{Key: "k"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.cfg)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if c != nil {
					t.Error("New() returned a non-nil client with an error")
				}
				return
			}
			if c == nil {
				t.Fatal("New() returned nil client without error")
			}
		})
	}
}

func TestNewDefaults(t *testing.T) {
	c, err := New(Config{Key: "k"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.key != "k" {
		t.Errorf("key = %q, want %q", c.key, "k")
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.usePackage {
		t.Error("usePackage should default to false (wallet)")
	}
	if c.httpClient == nil {
		t.Fatal("httpClient should default to a non-nil client")
	}
	if c.httpClient.Timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, DefaultTimeout)
	}
}

func TestNewOverrides(t *testing.T) {
	hc := &http.Client{Timeout: 3 * time.Second}
	c, err := New(Config{
		Key:        "k",
		UsePackage: true,
		BaseURL:    "https://private.gateway.example",
		HTTPClient: hc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.usePackage {
		t.Error("usePackage should be true")
	}
	if c.baseURL != "https://private.gateway.example" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if c.httpClient != hc {
		t.Error("httpClient not preserved")
	}
	if c.httpClient.Timeout != 3*time.Second {
		t.Errorf("timeout = %v, want 3s", c.httpClient.Timeout)
	}
}

func TestNewKeyNotTrimmed(t *testing.T) {
	// A key with surrounding whitespace is a different token. The client
	// must not silently mutate it.
	c, err := New(Config{Key: "  key  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.key != "  key  " {
		t.Errorf("key = %q, want %q", c.key, "  key  ")
	}
}
