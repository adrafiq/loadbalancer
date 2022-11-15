package main

import (
	"infrastructure/loadbalancer/proxy"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mocking http ResponseWriter
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type MockResponse struct {
	header int
}

func (r *MockResponse) Header() http.Header {
	return http.Header{}
}

func (r *MockResponse) Write([]byte) (int, error) {
	return 0, nil
}

func (r *MockResponse) WriteHeader(statusCode int) {
	r.header = statusCode
}

// Mocking io StringWriter
var contentWritten string

func writeString(w io.Writer, s string) (n int, err error) {
	contentWritten = s
	return 0, nil
}

func TestMakeHandler(t *testing.T) {
	request, _ := http.NewRequest("GET", "", nil)
	expectedName := "localhost"
	request.Host = expectedName
	host := proxy.Host{
		Name: expectedName,
	}
	var mockResponse MockResponse

	t.Run("responts 403 if host is different from the one configured", func(t *testing.T) {
		inputName := "different"
		expectedStatus := 403
		request.Host = inputName
		expectedMessage := "unrecognized host"
		handler := makeHandler(&host, writeString)
		contentWritten = ""
		handler(&mockResponse, request)
		if mockResponse.header != expectedStatus {
			t.Errorf("should have responded with expected status %d", expectedStatus)
		}
		if contentWritten != expectedMessage {
			t.Errorf("should have written content %s", expectedMessage)
		}
	})
	t.Run("responds 503 if there are no healthy servers", func(t *testing.T) {
		request.Host = expectedName
		host.HealthyServers = make([]proxy.Server, 0)
		handler := makeHandler(&host, writeString)
		expectedStatus := 503
		handler(&mockResponse, request)
		if mockResponse.header != expectedStatus {
			t.Errorf("should have responded with expected status %d", expectedStatus)
		}
	})
	t.Run("returns http response from selected upstream server in rotation", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", nil)
		req.Host = expectedName
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		serverAddress := strings.Split(server.URL, "//")[1]

		host = proxy.Host{
			Name:    "localhost",
			Scheme:  proxy.Random,
			Timeout: 10,
			HealthyServers: []proxy.Server{
				{Name: serverAddress},
			},
		}
		handler := makeHandler(&host, writeString)
		handler(&mockResponse, req)

		if mockResponse.header != http.StatusOK {
			t.Error("should have responded with expected status")
		}
	})
}
