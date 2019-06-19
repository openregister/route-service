package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPort               = "8080"
	defaultGoogleAnalyticsUrl = "https://www.google-analytics.com/collect"

	acceptHeader         = "Accept"
	userAgentHeader      = "User-Agent"
	cfForwardedURLHeader = "X-Cf-Forwarded-Url"
)

var (
	trackingID string

	timeout = time.Second * 5
)

type gaPayload struct {
	ProtocolVersion     int       `json:"v"`
	Type                string    `json:"t"`
	TrackingID          string    `json:"tid"`
	ClientID            uuid.UUID `json:"cid"`
	AnonymizeIP         bool      `json:"aip"`
	NonInteractionHit   bool      `json:"ni"`
	DocumentLocationURL string    `json:"dl"`
	UserAgent           string    `json:"ua"`
	APIKey              string    `json:"cd2"`
	ShortUserAgent      string    `json:"cd6"`
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(getLoggingLevel())

	if trackingID = os.Getenv("GOOGLE_ANALYTICS_TRACKING_ID"); len(trackingID) == 0 {
		log.Warn("no GOOGLE_ANALYTICS_TRACKING_ID set, disabling API calls")
	}

	skipSSLValidation := getSkipValidation()

	roundTripper := newDataCollectorRoundTripper(skipSSLValidation)
	proxy := newProxy(roundTripper, skipSSLValidation)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", getPort()), proxy))
}

func getPort() string {
	if port := os.Getenv("PORT"); len(port) != 0 {
		return port
	}

	return defaultPort
}

func getSkipValidation() bool {
	if os.Getenv("SKIP_SSL_VALIDATION") == "true" {
		return true
	}

	return false
}

func getLoggingLevel() log.Level {
	if os.Getenv("DEBUG") == "true" {
		return log.DebugLevel
	}

	return log.WarnLevel
}

func getGoogleAnalyticsUrl() string {
	if url := os.Getenv("GOOGLE_ANALYTICS_URL"); len(url) != 0 {
		return url
	}

	return defaultGoogleAnalyticsUrl
}

func newProxy(roundTripper http.RoundTripper, skipSSLValidation bool) http.Handler {
	reverseProxy := &httputil.ReverseProxy{
		Director:  reverseProxyDirectory,
		Transport: roundTripper,
	}

	return reverseProxy
}

func reverseProxyDirectory(req *http.Request) {
	forwardedURL := req.Header.Get(cfForwardedURLHeader)
	u, err := url.Parse(forwardedURL)
	if err != nil {
		log.Fatalln(err.Error())
	}

	req.URL = u
	req.Host = u.Host

	endpoint, err := composeResourceURL(req.Header, *req.URL)
	if err != nil {
		log.Errorf("unable to compose resource url: %s", err)
	}

	if len(trackingID) != 0 {
		go sendGARecord(endpoint, req.Header.Get(userAgentHeader))
	}
}

func composeResourceURL(headers http.Header, u url.URL) (string, error) {
	u.RawQuery = "" // TODO: Consider if this line needs to persist.

	log.WithFields(log.Fields{
		"accept": headers.Get(acceptHeader),
		"path":   u.Path,
	}).Debug("composing new resource url")

	accept := headers.Get(acceptHeader)

	switch {
	case strings.Contains(accept, "application/json"):
		u.Path = fmt.Sprintf("%s.json", u.Path)
		break
	case strings.Contains(accept, "text/csv"):
		u.Path = fmt.Sprintf("%s.csv", u.Path)
		break

	// following are not supported, yet we want to collect metrics on them
	case strings.Contains(accept, "text/html"):
		u.Path = fmt.Sprintf("%s.html", u.Path)
		break
	case strings.Contains(accept, "text/tab-separated-values"):
	case strings.Contains(accept, "text/tsv"):
		u.Path = fmt.Sprintf("%s.tsv", u.Path)
		break
	case strings.Contains(accept, "application/x-turtle"):
	case strings.Contains(accept, "text/ttl"):
		u.Path = fmt.Sprintf("%s.ttl", u.Path)
		break
	}

	ext := path.Ext(u.Path)

	switch ext {
	case "json":
	case "csv":
	case "html":
	case "tsv":
	case "ttl":
		break
	default:
		return "", fmt.Errorf("unknown resource type: %s", u.Path)
	}

	return u.String(), nil
}

func sendGARecord(endpoint string, useragent string) {
	log.WithField("endpoint", endpoint).Debug("sending google analytics record")
	data, err := json.Marshal(gaPayload{
		ProtocolVersion:     1,
		Type:                "pageview",
		TrackingID:          trackingID,
		ClientID:            uuid.New(),
		AnonymizeIP:         true,
		NonInteractionHit:   true,
		DocumentLocationURL: endpoint,
		UserAgent:           useragent,
	})
	if err != nil {
		log.Errorf("unable to compose google analytics request: %s", err)
		return
	}

	req, err := http.NewRequest("POST", getGoogleAnalyticsUrl(), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", string(len(data)))

	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("unable to perform google analytics request: %s", err)
	}
	defer resp.Body.Close()

	log.WithFields(log.Fields{
		"status": resp.StatusCode,
	}).Debug("received response from google analytics")
}

// DataCollectorRoundTripper is a struct holding some useful infromation and
// being an attachment point for functions fulfilling the RoundTripper
// interface.
type DataCollectorRoundTripper struct {
	transport http.RoundTripper
}

func newDataCollectorRoundTripper(skipSSLValidation bool) *DataCollectorRoundTripper {
	return &DataCollectorRoundTripper{
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipSSLValidation,
			},
		},
	}
}

// RoundTrip ...
func (rt *DataCollectorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := rt.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	log.WithField("url", req.URL.String()).Debug("sending response to gorouter")

	return res, err
}
