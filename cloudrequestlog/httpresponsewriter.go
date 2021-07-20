package cloudrequestlog

import "net/http"

type httpResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *httpResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (z *httpResponseWriter) Status() int {
	if z.statusCode == 0 {
		return http.StatusOK
	}
	return z.statusCode
}

func (z *httpResponseWriter) Write(b []byte) (int, error) {
	size, err := z.ResponseWriter.Write(b)
	z.size += size
	return size, err
}
