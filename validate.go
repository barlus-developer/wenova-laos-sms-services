package wenovasms

import (
	"regexp"
	"strings"
)

const (
	// MaxMessageLen is the maximum message length allowed by the Wenova API.
	MaxMessageLen = 500

	// DefaultHeaderOTP is the sender header used by SendOTP.
	DefaultHeaderOTP = "WNV-OTP"

	// DefaultHeaderInfo is the sender header used by SendInfo.
	DefaultHeaderInfo = "WNV-info"
)

var (
	phoneRe = regexp.MustCompile(`^20\d{8}$`)
	// linkRe matches URLs, domains, www., and short links. It is a
	// conservative heuristic; the Wenova API is the source of truth.
	linkRe = regexp.MustCompile(`(?i)(https?://|www\.|\.[a-z]{2,}\b)`)
)

// validatePhone reports whether phone is a Lao mobile number in
// 20XXXXXXXX format (10 digits starting with 20, no +856 or leading 0).
func validatePhone(phone string) bool {
	return phoneRe.MatchString(phone)
}

// validateMessage reports whether message is acceptable to the Wenova API:
// non-empty, at most MaxMessageLen runes, and free of links.
func validateMessage(message string) bool {
	if strings.TrimSpace(message) == "" {
		return false
	}
	if len([]rune(message)) > MaxMessageLen {
		return false
	}
	if linkRe.MatchString(message) {
		return false
	}
	return true
}
