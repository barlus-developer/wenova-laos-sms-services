package wenovasms

import (
	"strings"
	"testing"
)

func TestValidatePhone(t *testing.T) {
	valid := []string{
		"2000000000",
		"2012345678",
		"2099999999",
		"2022222222",
	}
	invalid := []string{
		"",
		"1234567890",     // not starting with 20
		"201234567",      // 9 digits
		"20123456789",    // 11 digits
		"+8562012345678", // includes +856
		"8562012345678",  // includes country code without +
		"02012345678",    // leading 0
		"201234567a",     // non-digit
		" 2012345678",    // leading space
		"2012345678 ",    // trailing space
		"20-1234-5678",   // separators
		"2012 345678",    // internal space
		"20123X56789",    // non-digit in middle
	}

	for _, p := range valid {
		if !validatePhone(p) {
			t.Errorf("validatePhone(%q) = false, want true", p)
		}
	}
	for _, p := range invalid {
		if validatePhone(p) {
			t.Errorf("validatePhone(%q) = true, want false", p)
		}
	}
}

func TestValidateMessage(t *testing.T) {
	t.Run("valid messages", func(t *testing.T) {
		valid := []string{
			"Your OTP is: 123456",
			"ຂໍ້ຄວາມພາສາລາວ",
			"a", // single char
			strings.Repeat("a", MaxMessageLen),
		}
		for _, m := range valid {
			if !validateMessage(m) {
				t.Errorf("validateMessage(%q) = false, want true", m)
			}
		}
	})

	t.Run("empty and too long", func(t *testing.T) {
		invalid := []string{
			"",
			"   ",
			strings.Repeat("a", MaxMessageLen+1),
		}
		for _, m := range invalid {
			if validateMessage(m) {
				t.Errorf("validateMessage(%q) = true, want false", m)
			}
		}
	})

	t.Run("unicode rune count not byte count", func(t *testing.T) {
		lao := "ກ"
		msg := strings.Repeat(lao, MaxMessageLen) // 500 runes, 1500 bytes
		if !validateMessage(msg) {
			t.Error("500 Lao runes should be valid (length is in runes, not bytes)")
		}
		if validateMessage(strings.Repeat(lao, MaxMessageLen+1)) {
			t.Error("501 Lao runes should be invalid")
		}
	})

	t.Run("links rejected", func(t *testing.T) {
		links := []string{
			"visit https://example.com",
			"visit http://example.com",
			"see www.example.com",
			"go to example.com now",
			"short link bit.ly/xyz",
			"use api.github.com",
			"https://www.example.com/path?q=1",
		}
		for _, m := range links {
			if validateMessage(m) {
				t.Errorf("validateMessage(%q) = true, want false (link)", m)
			}
		}
	})

	t.Run("non-link dots allowed", func(t *testing.T) {
		valid := []string{
			"Mr. Smith please call us",
			"Total: 500.00 LAK",
			"version 1.2 is ready",
			"meet at 9 a.m.",
		}
		for _, m := range valid {
			if !validateMessage(m) {
				t.Errorf("validateMessage(%q) = false, want true (not a link)", m)
			}
		}
	})
}

func TestValidateHeaderDefaults(t *testing.T) {
	if DefaultHeaderOTP != "WNV-OTP" {
		t.Errorf("DefaultHeaderOTP = %q, want WNV-OTP", DefaultHeaderOTP)
	}
	if DefaultHeaderInfo != "WNV-info" {
		t.Errorf("DefaultHeaderInfo = %q, want WNV-info", DefaultHeaderInfo)
	}
}
