package wenovasms

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// newTestClient spins up an httptest server and returns a Client wired
// to it, so tests exercise the full HTTP path without touching the
// network.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(Config{
		Key:        "test-key",
		UsePackage: true,
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

type requestBody struct {
	Header      string `json:"header"`
	PhoneNumber string `json:"phoneNumber"`
	Message     string `json:"message"`
	Token       string `json:"token"`
	UsePackage  bool   `json:"usePackage"`
}

func TestSendMethodsBuildRequest(t *testing.T) {
	tests := []struct {
		name    string
		send    func(*Client) (*SendResult, error)
		wantHdr string
		usePkg  bool
	}{
		{
			name: "SendOTP",
			send: func(c *Client) (*SendResult, error) {
				return c.SendOTP(context.Background(), "2012345678", "Your OTP is: 123456")
			},
			wantHdr: DefaultHeaderOTP,
			usePkg:  true,
		},
		{
			name: "SendInfo",
			send: func(c *Client) (*SendResult, error) {
				return c.SendInfo(context.Background(), "2012345678", "Welcome!")
			},
			wantHdr: DefaultHeaderInfo,
			usePkg:  true,
		},
		{
			name: "SendCustom",
			send: func(c *Client) (*SendResult, error) {
				return c.SendCustom(context.Background(), "WNV-Promo", "2012345678", "50% off!")
			},
			wantHdr: "WNV-Promo",
			usePkg:  true,
		},
		{
			name: "SendCustomWallet",
			send: func(c *Client) (*SendResult, error) {
				return c.SendCustom(context.Background(), "WNV-Txn", "2012345678", "Payment received")
			},
			wantHdr: "WNV-Txn",
			usePkg:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the server first so we can configure usePackage per case.
			var body requestBody
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != apiPath {
					t.Errorf("path = %q, want %q", r.URL.Path, apiPath)
				}
				if r.Method != http.MethodPost {
					t.Errorf("method = %q, want POST", r.Method)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode request: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":{"resultCode":"20000","resultDesc":"success"}}`)
			}))
			t.Cleanup(srv.Close)

			c, err := New(Config{
				Key:        "test-key",
				UsePackage: tt.usePkg,
				BaseURL:    srv.URL,
				HTTPClient: srv.Client(),
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			res, err := tt.send(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.ResultCode != "20000" {
				t.Errorf("ResultCode = %q, want 20000", res.ResultCode)
			}
			if res.ResultDesc != "success" {
				t.Errorf("ResultDesc = %q, want success", res.ResultDesc)
			}
			if body.Header != tt.wantHdr {
				t.Errorf("header = %q, want %q", body.Header, tt.wantHdr)
			}
			if body.PhoneNumber != "2012345678" {
				t.Errorf("phoneNumber = %q", body.PhoneNumber)
			}
			if body.Token != "test-key" {
				t.Errorf("token = %q", body.Token)
			}
			if body.UsePackage != tt.usePkg {
				t.Errorf("usePackage = %v, want %v", body.UsePackage, tt.usePkg)
			}
			if !strings.Contains(string(res.Data), "resultCode") {
				t.Errorf("Data should carry the raw payload, got %s", res.Data)
			}
		})
	}
}

func TestSendValidationErrors(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for invalid input")
	})

	if _, err := c.SendOTP(context.Background(), "bad-phone", "msg"); !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("want ErrInvalidPhone, got %v", err)
	}
	if _, err := c.SendCustom(context.Background(), "", "2012345678", "msg"); !errors.Is(err, ErrInvalidHeader) {
		t.Fatalf("want ErrInvalidHeader, got %v", err)
	}
	if _, err := c.SendInfo(context.Background(), "2012345678", "visit example.com"); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("want ErrInvalidMessage, got %v", err)
	}
}

func TestSendBusinessError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":30105,"message":"SMS package quota is insufficient","path":"/sms/package","method":"POST"}`)
	})

	_, err := c.SendOTP(context.Background(), "2012345678", "msg")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != CodePackageInsufficient {
		t.Errorf("Code = %d, want %d", apiErr.Code, CodePackageInsufficient)
	}
	if apiErr.HTTP != http.StatusBadRequest {
		t.Errorf("HTTP = %d, want %d", apiErr.HTTP, http.StatusBadRequest)
	}
	if apiErr.Message == "" {
		t.Error("Message should be parsed")
	}
	if strings.Contains(err.Error(), "test-key") {
		t.Error("error must not leak the api key")
	}
}

func TestSendEmptyData(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})

	res, err := c.SendOTP(context.Background(), "2012345678", "msg")
	if err == nil {
		t.Fatal("expected GatewayError for empty resultCode")
	}
	var gwErr *GatewayError
	if !errors.As(err, &gwErr) {
		t.Fatalf("expected *GatewayError, got %T", err)
	}
	if res == nil {
		t.Fatal("expected a result even when data is empty")
	}
	if len(res.Data) != 0 {
		t.Errorf("Data should be empty, got %s", res.Data)
	}
}

func TestSendGatewayError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"resultCode":"30001","resultDesc":"rejected"}}`)
	})

	res, err := c.SendOTP(context.Background(), "2012345678", "msg")
	if err == nil {
		t.Fatal("expected error")
	}
	var gwErr *GatewayError
	if !errors.As(err, &gwErr) {
		t.Fatalf("expected *GatewayError, got %T", err)
	}
	if gwErr.ResultCode != "30001" || gwErr.ResultDesc != "rejected" {
		t.Errorf("GatewayError = %+v", gwErr)
	}
	if res == nil || res.ResultCode != "30001" {
		t.Errorf("result should carry the gateway code, got %+v", res)
	}
}

func TestSendMalformedErrorBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `this is not json`)
	})

	_, err := c.SendOTP(context.Background(), "2012345678", "msg")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.HTTP != http.StatusServiceUnavailable {
		t.Errorf("HTTP = %d, want %d", apiErr.HTTP, http.StatusServiceUnavailable)
	}
}

func TestSendMalformedSuccessBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `this is not json`)
	})

	_, err := c.SendOTP(context.Background(), "2012345678", "msg")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected a JSON parse error, got %v", err)
	}
}

func TestSendServerFiveHundred(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := c.SendOTP(context.Background(), "2012345678", "msg")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.HTTP != http.StatusInternalServerError {
		t.Errorf("HTTP = %d, want %d", apiErr.HTTP, http.StatusInternalServerError)
	}
}

func TestSendNetworkError(t *testing.T) {
	// Point at a closed server to force a connection error.
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	c, err := New(Config{Key: "k", BaseURL: url, HTTPClient: &http.Client{Timeout: 2 * time.Second}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = c.SendOTP(context.Background(), "2012345678", "msg")
	if err == nil {
		t.Fatal("expected network error")
	}
	var apiErr *Error
	if errors.As(err, &apiErr) {
		t.Error("network errors should not be wrapped as API errors")
	}
}

func TestSendContextCancelled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = io.WriteString(w, `{"data":{"resultCode":"20000","resultDesc":"success"}}`)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := c.SendOTP(ctx, "2012345678", "msg")
	if err == nil {
		t.Fatal("expected context deadline error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestSendAlreadyCancelled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for a cancelled context")
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.SendOTP(ctx, "2012345678", "msg")
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSendConcurrent(t *testing.T) {
	const n = 50

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&requestBody{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"resultCode":"20000","resultDesc":"success"}}`)
	})

	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = c.SendOTP(context.Background(), "2012345678", "concurrent message")
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("request %d failed: %v", i, err)
		}
	}
}
