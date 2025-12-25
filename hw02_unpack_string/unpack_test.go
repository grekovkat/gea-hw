package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{"3abc", "45", "aaa10b"}
	for _, tc := range invalidStrings {
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}

func TestIsCorrect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "digit then digit", input: "22", expected: false},
		{name: "slash start", input: `\2`, expected: true},
		{name: "slash char", input: `\a`, expected: false},
		{name: "slash unicode", input: `\🙃`, expected: false},
		{name: "slash unicode", input: `a2`, expected: true},
		{name: "slash unicode", input: `🙃2`, expected: true},
		{name: "unicode unicode", input: `🙃🙃`, expected: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runes := []rune(tc.input)
			result := IsCorrect(runes[0], runes[1])
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackNoPanic(t *testing.T) {
	trickyCases := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "single char", input: "a"},
		{name: "single slash", input: `\`},
		{name: "start with digit", input: "0"},
		{name: "single unicode", input: "🙃"},
	}

	for _, tc := range trickyCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, _ = Unpack(tc.input)
			})

			result, err := Unpack(tc.input)
			t.Logf("Input: %q, Result: %q, Error: %v", tc.input, result, err)
		})
	}
}

func TestValidateRunes(t *testing.T) {
	testStrings := []string{"3abc", "aaa10b", `a2\`}

	for _, input := range testStrings {
		t.Run(input, func(t *testing.T) {
			err := ValidateRunes([]rune(input))
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error: %q", err)
		})
	}
}
