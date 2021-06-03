package shuttle

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	assertions := []struct {
		Input  interface{}
		Accept string
		HTTPResponse
	}{
		{Input: nil,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: nil, Body: ""}},
		{Input: "body",
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: []byte("body"),
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},

		{Input: true,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "true"}},
		{Input: false,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "false"}},

		{Input: &TextResult{StatusCode: 201, ContentType: "no-expected-content-type-for-empty-body", Content: ""},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: nil, Body: ""}},
		{Input: TextResult{StatusCode: 201, ContentType: "no-expected-content-type-for-empty-body", Content: ""},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: nil, Body: ""}},
		{Input: &TextResult{StatusCode: 0, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: &TextResult{StatusCode: 201, ContentType: "application/custom", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: TextResult{StatusCode: 0, ContentType: "application/custom", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: TextResult{StatusCode: 202, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 202, ContentType: nil, Body: "body"}},

		{Input: &BinaryResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: BinaryResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: &BinaryResult{StatusCode: 404, ContentType: "", Content: []byte("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},
		{Input: BinaryResult{StatusCode: 404, ContentType: "", Content: []byte("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},

		{Input: &StreamResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: StreamResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: &StreamResult{StatusCode: 404, ContentType: "", Content: bytes.NewBufferString("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},
		{Input: StreamResult{StatusCode: 422, ContentType: "application/custom", Content: bytes.NewBufferString("body")},
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/custom"}, Body: "body"}},

		{Input: &SerializeResult{StatusCode: 401, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 401, ContentType: nil, Body: ""}},
		{Input: SerializeResult{StatusCode: 401, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 401, ContentType: nil, Body: ""}},

		{Input: &SerializeResult{StatusCode: 422, ContentType: "", Content: "body"},
			Accept:       "", // default serializer
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/json; charset=utf-8"}, Body: "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "application/override-default", Content: "body"},
			Accept:       "", // default serializer
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/override-default"}, Body: "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "", Content: "body"},
			Accept:       "application/xml", // serializer matching this Accept value
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/xml; charset=utf-8"}, Body: "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "application/override-default", Content: "body"},
			Accept:       "application/xml", // serializer matching this Accept value
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/override-default"}, Body: "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "", Content: "body"},
			Accept:       "application/not-acceptable", // use default serializer
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/json; charset=utf-8"}, Body: "{body}"}},

		{Input: &SerializeResult{StatusCode: 200, ContentType: "", Content: "body"},
			Accept:       "application/xml;q=0.8", // simplify and use correct serializer
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"application/xml; charset=utf-8"}, Body: "{body}"}},

		{Input: 42, // use serializer for unknown type
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"application/json; charset=utf-8"}, Body: "{42}"}},
	}

	for _, assertion := range assertions {
		response := recordResponse(assertion.Input, assertion.Accept)
		assertResponse(t, response, assertion.HTTPResponse)
	}
}
func recordResponse(result interface{}, acceptHeader string) *httptest.ResponseRecorder {
	writer := newTestWriter()
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	if len(acceptHeader) > 0 {
		request.Header["Accept"] = []string{acceptHeader}
	}

	writer.Write(response, request, result)

	return response
}
func newTestWriter() Writer {
	return newWriter(map[string]func() Serializer{
		"":                func() Serializer { return newTestWriteSerializer("application/json; charset=utf-8") },
		"application/xml": func() Serializer { return newTestWriteSerializer("application/xml; charset=utf-8") },
	})
}
func assertResponse(t *testing.T, response *httptest.ResponseRecorder, expected HTTPResponse) {
	Assert(t).That(response.Code).Equals(expected.StatusCode)
	Assert(t).That(response.Header()["Content-Type"]).Equals(expected.ContentType)
	Assert(t).That(response.Body.String()).Equals(expected.Body)
}

func TestWriteHTTPHandler(t *testing.T) {
	handler := &TestHTTPHandlerResult{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	newTestWriter().Write(response, request, handler)

	Assert(t).That(handler.response == response).IsTrue()
	Assert(t).That(handler.request == request).IsTrue()
}

type HTTPResponse struct {
	StatusCode  int
	ContentType []string
	Body        string
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TestHTTPHandlerResult struct {
	response http.ResponseWriter
	request  *http.Request
}

func (this *TestHTTPHandlerResult) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.response = response
	this.request = request
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TestWriteSerializer string

func newTestWriteSerializer(contentType string) Serializer {
	return TestWriteSerializer(contentType)
}
func (this TestWriteSerializer) ContentType() string { return string(this) }
func (this TestWriteSerializer) Serialize(writer io.Writer, value interface{}) error {
	raw, _ := json.Marshal(value)
	_, _ = io.WriteString(writer, "{"+strings.ReplaceAll(string(raw), `"`, ``)+"}")
	return nil
}
