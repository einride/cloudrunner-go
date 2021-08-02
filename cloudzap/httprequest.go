package cloudzap

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// HTTPRequest creates a new zap.Field for a Cloud Logging HTTP request.
// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
func HTTPRequest(r *HTTPRequestObject) zap.Field {
	return zap.Object("httpRequest", r)
}

// HTTPRequestObject is a common message for logging HTTP requests.
// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
type HTTPRequestObject struct {
	// RequestMethod is the request method. Examples: "GET", "HEAD", "PUT", "POST".
	RequestMethod string
	// RequestURL is the scheme (http, https), the host name, the path and the query portion of the URL
	// that was requested. Example: "http://example.com/some/info?color=red".
	RequestURL string
	// The size of the HTTP request message in bytes, including the request headers and the request body.
	RequestSize int
	// Status is the response code indicating the status of response. Examples: 200, 404.
	Status int
	// ResponseSize is the size of the HTTP response message sent back to the client, in bytes, including the response headers
	// and the response body.
	ResponseSize int
	// UserAgent is the user agent sent by the client.
	// Example: "Mozilla/4.0 (compatible; MSIE 6.0; Windows 98; Q312461; .NET CLR 1.0.3705)".
	UserAgent string
	// RemoteIP is the IP address (IPv4 or IPv6) of the client that issued the HTTP request.
	// This field can include port information.
	// Examples: "192.168.1.1", "10.0.0.1:80", "FE80::0202:B3FF:FE1E:8329".
	RemoteIP string
	// ServerIP is the IP address (IPv4 or IPv6) of the origin server that the request was sent to.
	// This field can include port information.
	// Examples: "192.168.1.1", "10.0.0.1:80", "FE80::0202:B3FF:FE1E:8329".
	ServerIP string
	// Referer is the referer URL of the request, as defined in HTTP/1.1 Header Field Definitions.
	Referer string
	// Latency is the request processing latency on the server, from the time the request was received
	// until the response was sent.
	Latency time.Duration
	// Protocol is the protocol used for the request. Examples: "HTTP/1.1", "HTTP/2", "websocket"
	Protocol string
}

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (h *HTTPRequestObject) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	if h.RequestMethod != "" {
		encoder.AddString("requestMethod", h.RequestMethod)
	}
	if h.RequestURL != "" {
		encoder.AddString("requestUrl", fixUTF8(h.RequestURL))
	}
	if h.RequestSize > 0 {
		addInt(encoder, "requestSize", h.RequestSize)
	}
	if h.Status != 0 {
		encoder.AddInt("status", h.Status)
	}
	if h.ResponseSize > 0 {
		addInt(encoder, "responseSize", h.ResponseSize)
	}
	if h.UserAgent != "" {
		encoder.AddString("userAgent", h.UserAgent)
	}
	if h.RemoteIP != "" {
		encoder.AddString("remoteIp", h.RemoteIP)
	}
	if h.ServerIP != "" {
		encoder.AddString("serverIp", h.ServerIP)
	}
	if h.Referer != "" {
		encoder.AddString("referer", h.Referer)
	}
	if h.Latency > 0 {
		addDuration(encoder, "latency", h.Latency)
	}
	if h.Protocol != "" {
		encoder.AddString("protocol", h.Protocol)
	}
	return nil
}

func addDuration(encoder zapcore.ObjectEncoder, key string, d time.Duration) {
	// A duration in seconds with up to nine fractional digits, terminated by 's'. Example: "3.5s".
	encoder.AddString(key, fmt.Sprintf("%fs", d.Seconds()))
}

func addInt(encoder zapcore.ObjectEncoder, key string, i int) {
	encoder.AddString(key, strconv.Itoa(i))
}

// fixUTF8 is copied from cloud.google.com/logging/internal and fixes invalid UTF-8 strings.
// See: https://github.com/googleapis/google-cloud-go/issues/1383
func fixUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var buf strings.Builder
	buf.Grow(len(s))
	for _, r := range s {
		if utf8.ValidRune(r) {
			buf.WriteRune(r)
		} else {
			buf.WriteRune('\uFFFD')
		}
	}
	return buf.String()
}
