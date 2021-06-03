package shuttle

import "io"

// TextResult provides the ability render a result which contains text.
type TextResult struct {
	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int
	// ContentType, if provided, use this value.
	ContentType string
	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content string
}

// BinaryResult provides the ability render a result which contains binary data.
type BinaryResult struct {
	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int
	// ContentType, if provided, use this value.
	ContentType string
	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content []byte
}

type StreamResult struct {
	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int
	// ContentType, if provided, use this value.
	ContentType string
	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content io.Reader
}

type SerializeResult struct {
	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int
	// ContentType, if provided, use this value.
	ContentType string
	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content interface{}
}

// InputError represents some kind of problem with the calling HTTP request.
type InputError struct {
	// Fields indicates the exact location(s) of the errors including the part of the HTTP request itself this is
	// invalid. Valid field prefixes include "path", "query", "header", "form", and "body".
	Fields []string `json:"fields,omitempty"`
	// ID represents the unique, numeric contractual identifier that can be used to associate this error with a particular front-end error message, if any.
	ID int `json:"id,omitempty"`
	// Name represents the unique string-based, contractual value that can be used to associate this error with a particular front-end error message, if any.
	Name string `json:"name,omitempty"`
	// Message represents a friendly, user-facing message to indicate why there was a problem with the input.
	Message string `json:"message,omitempty"`
}

func (this InputError) Error() string { return this.Message }