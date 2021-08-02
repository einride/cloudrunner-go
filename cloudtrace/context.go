package cloudtrace

import (
	"fmt"
	"strings"
)

// Context represents a Google Cloud Trace context header value.
//
// The format of the X-Cloud-Trace-Context header is:
//
//  TRACE_ID/SPAN_ID;o=TRACE_TRUE"
//
// See: https://cloud.google.com/trace/docs/setup
type Context struct {
	// TraceID is a 32-character hexadecimal value representing a 128-bit number.
	TraceID string
	// SpanID is the decimal representation of the (unsigned) span ID.
	SpanID string
	// Sampled indicates if the trace is being sampled.
	Sampled bool
}

// UnmarshalString parses the provided X-Cloud-Trace-Context header.
func (c *Context) UnmarshalString(value string) error {
	if len(value) == 0 {
		return fmt.Errorf("empty %s", ContextHeader)
	}
	indexOfSlash := strings.IndexByte(value, '/')
	if indexOfSlash == -1 {
		c.TraceID = value
		if len(c.TraceID) != 32 {
			return fmt.Errorf("invalid %s '%s': trace ID is not a 32-character hex value", ContextHeader, value)
		}
		return nil
	}
	c.TraceID = value[:indexOfSlash]
	if len(c.TraceID) != 32 {
		return fmt.Errorf("invalid %s '%s': trace ID is not a 32-character hex value", ContextHeader, value)
	}
	indexOfSemicolon := strings.IndexByte(value, ';')
	if indexOfSemicolon == -1 {
		c.SpanID = value[indexOfSlash+1:]
		return nil
	}
	c.SpanID = value[indexOfSlash+1 : indexOfSemicolon]
	switch value[indexOfSemicolon+1:] {
	case "o=1":
		c.Sampled = true
	case "o=0":
		c.Sampled = false
	default:
		return fmt.Errorf("invalid %s '%s'", ContextHeader, value)
	}
	return nil
}

// String returns a string representation of the trace context.
func (c Context) String() string {
	sampled := 0
	if c.Sampled {
		sampled = 1
	}
	return fmt.Sprintf("%s/%s;o=%d", c.TraceID, c.SpanID, sampled)
}
