package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

const (
	host        = "example.com"
	httpScheme  = "http"
	httpsScheme = "https"

	endpointJSON = "https://example.com/resource.json"
	endpointHTML = "https://example.com/resource.html"
	endpointCSV  = "https://example.com/resource.csv"
)

var (
	headers = http.Header{}
)

func TestComposeResourceURL(t *testing.T) {
	acceptHTML := http.Header{}
	acceptJSON := http.Header{}
	acceptCSV := http.Header{}

	acceptJSON.Set("Accept", "application/json")
	acceptHTML.Set("Accept", "text/html")
	acceptCSV.Set("Accept", "text/csv")

	var cases = []struct {
		url     url.URL
		headers http.Header
		out     string
	}{
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource.json"}, headers, endpointJSON},
		{url.URL{Scheme: httpScheme, Host: host, Path: "resource.json"}, headers, "http://example.com/resource.json"},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource.csv"}, headers, endpointCSV},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource"}, acceptJSON, endpointJSON},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource"}, acceptHTML, endpointHTML},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource"}, acceptCSV, endpointCSV},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource", RawQuery: "sort=name"}, acceptJSON, endpointJSON},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource", RawQuery: "sort=name"}, acceptCSV, endpointCSV},
	}

	for _, testCase := range cases {
		t.Run(fmt.Sprintf("endpoint=%s?%s#accept_header=%s", testCase.url.Path, testCase.url.RawQuery, testCase.headers.Get(acceptHeader)), func(t *testing.T) {
			endpoint, err := composeResourceURL(testCase.headers, testCase.url)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if endpoint != testCase.out {
				t.Errorf("got %s, want %s", endpoint, testCase.out)
			}
		})
	}
}

func TestComposeResourceURLFailure(t *testing.T) {
	acceptPNG := http.Header{}
	acceptPNG.Set("Accept", "image/png")

	var cases = []struct {
		url     url.URL
		headers http.Header
		out     string
	}{
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource"}, headers, ""},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource.png"}, headers, ""},
		{url.URL{Scheme: httpsScheme, Host: host, Path: "resource"}, acceptPNG, ""},
	}

	for _, testCase := range cases {
		t.Run(fmt.Sprintf("endpoint=%s?%s#accept_header=%s", testCase.url.Path, testCase.url.RawQuery, testCase.headers.Get(acceptHeader)), func(t *testing.T) {
			endpoint, err := composeResourceURL(testCase.headers, testCase.url)
			if err == nil {
				t.Errorf("expected error to have occurred")
			}
			if endpoint != testCase.out {
				t.Errorf("expected no endpoint to be returned, got: %s", endpoint)
			}
		})
	}
}
