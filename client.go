package wenovasms

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// apiPath is the Wenova endpoint for sending SMS.
const apiPath = "/sms/package"

// Client sends SMS through the Wenova Link API.
//
// Create one with New. A Client is safe for concurrent use.
type Client struct {
	key        string
	usePackage bool
	baseURL    string
	httpClient *http.Client
}

// SendResult reports the SMS gateway outcome for a single send.
type SendResult struct {
	// ResultCode is the SMS gateway result code. "20000" means success.
	ResultCode string
	// ResultDesc is the SMS gateway result description.
	ResultDesc string
	// Data is the raw "data" payload from the API response, for
	// inspecting additional gateway fields.
	Data json.RawMessage
}

// SendOTP sends an SMS with the OTP sender header (DefaultHeaderOTP).
func (c *Client) SendOTP(ctx context.Context, phoneNumber, message string) (*SendResult, error) {
	return c.send(ctx, DefaultHeaderOTP, phoneNumber, message)
}

// SendInfo sends an SMS with the info sender header (DefaultHeaderInfo).
func (c *Client) SendInfo(ctx context.Context, phoneNumber, message string) (*SendResult, error) {
	return c.send(ctx, DefaultHeaderInfo, phoneNumber, message)
}

// SendCustom sends an SMS with a custom registered sender header.
func (c *Client) SendCustom(ctx context.Context, header, phoneNumber, message string) (*SendResult, error) {
	return c.send(ctx, header, phoneNumber, message)
}

// send validates the inputs, builds the request, and decodes the
// Wenova response. The API key is placed only in the request body and
// is never included in errors or logs.
func (c *Client) send(ctx context.Context, header, phoneNumber, message string) (*SendResult, error) {
	if header == "" {
		return nil, ErrInvalidHeader
	}
	if !validatePhone(phoneNumber) {
		return nil, ErrInvalidPhone
	}
	if !validateMessage(message) {
		return nil, ErrInvalidMessage
	}

	body, err := json.Marshal(sendRequest{
		Header:      header,
		PhoneNumber: phoneNumber,
		Message:     message,
		Token:       c.key,
		UsePackage:  c.usePackage,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+apiPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, decodeAPIError(resp.StatusCode, raw)
	}

	var out struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}

	var data struct {
		ResultCode string `json:"resultCode"`
		ResultDesc string `json:"resultDesc"`
	}
	if len(out.Data) > 0 {
		_ = json.Unmarshal(out.Data, &data)
	}

	result := &SendResult{
		ResultCode: data.ResultCode,
		ResultDesc: data.ResultDesc,
		Data:       out.Data,
	}

	// A successful SMS normally has resultCode "20000". Anything else —
	// including an empty result (still processing) — is reported as a
	// gateway error so the caller can inspect it.
	if result.ResultCode != "20000" {
		return result, &GatewayError{
			ResultCode: result.ResultCode,
			ResultDesc: result.ResultDesc,
		}
	}
	return result, nil
}

type sendRequest struct {
	Header      string `json:"header"`
	PhoneNumber string `json:"phoneNumber"`
	Message     string `json:"message"`
	Token       string `json:"token"`
	UsePackage  bool   `json:"usePackage"`
}

// decodeAPIError converts a non-2xx Wenova response into an *Error.
func decodeAPIError(status int, raw []byte) error {
	var apiErr struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Path    string `json:"path"`
		Method  string `json:"method"`
	}
	// If the body is not the expected shape, still report the status.
	_ = json.Unmarshal(raw, &apiErr)
	return &Error{
		Code:    Code(apiErr.Code),
		Message: apiErr.Message,
		Path:    apiErr.Path,
		Method:  apiErr.Method,
		HTTP:    status,
	}
}
