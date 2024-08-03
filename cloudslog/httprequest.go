package cloudslog

import (
	"bytes"
	"unicode/utf8"

	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/proto"
)

func fixHTTPRequest(r *ltype.HttpRequest) *ltype.HttpRequest {
	// Fix issue with invalid UTF-8.
	// See: https://github.com/googleapis/google-cloud-go/issues/1383.
	if fixedRequestURL := fixUTF8(r.RequestUrl); fixedRequestURL != r.RequestUrl {
		r = proto.Clone(r).(*ltype.HttpRequest)
		r.RequestUrl = fixedRequestURL
	}
	return r
}

// fixUTF8 is a helper that fixes an invalid UTF-8 string by replacing
// invalid UTF-8 runes with the Unicode replacement character (U+FFFD).
// See: https://github.com/googleapis/google-cloud-go/issues/1383.
func fixUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// Otherwise time to build the sequence.
	buf := new(bytes.Buffer)
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
