package cloudserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.einride.tech/cloudrunner/cloudserver"
	"gotest.tools/v3/assert"
)

func TestHTTPServer_RescuePanicsWithStatusInternalServerError(t *testing.T) {
	var handler http.Handler = http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom!")
	})
	res := serveHTTP(handler)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestHTTPServer_WorksWhenHeaderIsAlreadyWritten(t *testing.T) {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		panic("boom!")
	})
	res := serveHTTP(handler)
	assert.Equal(t, res.Code, http.StatusOK)
}

func serveHTTP(handler http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
	res := httptest.NewRecorder()

	middleware := cloudserver.Middleware{}
	handler = middleware.HTTPServer(handler)
	handler.ServeHTTP(res, req)

	return res
}
