package wenovasms

import (
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL is the Wenova Link SMS API endpoint.
const DefaultBaseURL = "https://apimicroservices.wenova.fun"

// DefaultTimeout is used when Config.HTTPClient does not set its own
// Timeout.
const DefaultTimeout = 10 * time.Second

// Config configures the Wenova SMS client.
//
// Only Key is required. All other fields are optional and fall back to
// sensible defaults when left at their zero value.
type Config struct {
	// Key is the Wenova API token from Dashboard → Key. Required.
	// It is kept private on the Client and never included in errors or
	// logs.
	Key string

	// UsePackage selects the payment source:
	//   true  — SMS is deducted from your SMS package quota
	//   false — SMS is charged from your wallet (default)
	UsePackage bool

	// BaseURL overrides the Wenova API endpoint. Intended for testing
	// and private gateways. Defaults to DefaultBaseURL.
	BaseURL string

	// HTTPClient overrides the HTTP client used for requests. Defaults
	// to a client with a DefaultTimeout timeout.
	HTTPClient *http.Client
}

// New validates cfg and returns a ready-to-use Client.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.Key) == "" {
		return nil, ErrInvalidKey
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultTimeout}
	}

	return &Client{
		key:        cfg.Key,
		usePackage: cfg.UsePackage,
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}
