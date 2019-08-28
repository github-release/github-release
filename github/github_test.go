package github

import (
	"net/url"
	"testing"
)

func TestRedactedURL(t *testing.T) {
	type testCase struct {
		name     string
		initial  *url.URL
		expected *url.URL
	}

	testCases := []testCase{
		testCase{
			name: "without access_token query",
			initial: &url.URL{
				Host:     "example.com",
				Scheme:   "https",
				RawQuery: "foo=bar",
			},
			expected: &url.URL{
				Host:     "example.com",
				Scheme:   "https",
				RawQuery: "foo=bar",
			},
		},
		testCase{
			name: "with access_token query",
			initial: &url.URL{
				Host:     "example.com",
				Scheme:   "https",
				RawQuery: "access_token=secret&foo=bar",
			},
			expected: &url.URL{
				Host:     "example.com",
				Scheme:   "https",
				RawQuery: "access_token=REDACTED&foo=bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := redactedURL(tc.initial)
			if tc.expected.String() != gotURL.String() {
				t.Fatalf("URLs do not match; want: %s, got %s", tc.expected.String(), gotURL.String())
			}
		})
	}
}
