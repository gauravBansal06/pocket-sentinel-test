package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"
	gohttp "github.com/go-kit/kit/transport/http"
)

// DecodeRequestFunc extracts a user-domain request object from an HTTP
// request object. Copied from gokit
type DecodeRequestFunc func(context.Context, *gin.Context, lumber.Logger) (request interface{}, err error)

// EncodeResponseFunc encodes the passed response object to the HTTP response
// writer. Copied from go-kit
type EncodeResponseFunc func(context.Context, *gin.Context, interface{}, lumber.Logger) error

// NopRequestDecoder is a DecodeRequestFunc that can be used for requests that do not
// need to be decoded, and simply returns nil, nil.
func NopRequestDecoder(ctx context.Context, c *gin.Context, logger lumber.Logger) (interface{}, error) {
	return nil, nil
}

// EncodeJSONResponse is a EncodeResponseFunc that serializes the response as a
// JSON object to the ResponseWriter. Copied from gokit
func EncodeJSONResponse(_ context.Context, c *gin.Context, response interface{}, logger lumber.Logger) error {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	if headerer, ok := response.(gohttp.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				c.Writer.Header().Add(k, v)
			}
		}
	}
	code := http.StatusOK
	if sc, ok := response.(gohttp.StatusCoder); ok {
		code = sc.StatusCode()
	}
	c.Writer.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}
	return json.NewEncoder(c.Writer).Encode(response)
}

// DefaultErrorEncoder writes the error to the ResponseWriter, by default a
// content type of text/plain, a body of the plain text of the error, and a
// status code of 500. Copied from gokit
func DefaultErrorEncoder(_ context.Context, err error, c *gin.Context, logger lumber.Logger) {
	contentType, body := "text/plain; charset=utf-8", []byte(err.Error())
	if marshaler, ok := err.(json.Marshaler); ok {
		if jsonBody, marshalErr := marshaler.MarshalJSON(); marshalErr == nil {
			contentType, body = "application/json; charset=utf-8", jsonBody
		}
	}
	c.Writer.Header().Set("Content-Type", contentType)
	if headerer, ok := err.(gohttp.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				c.Writer.Header().Add(k, v)
			}
		}
	}
	code := http.StatusInternalServerError
	if sc, ok := err.(gohttp.StatusCoder); ok {
		code = sc.StatusCode()
	}
	c.Writer.WriteHeader(code)
	c.Writer.Write(body)
}

// NewHTTPHandler generates gin handler from an endpoint and decoder/encoder
func NewHTTPHandler(ep endpoint.Endpoint, dec DecodeRequestFunc, enc EncodeResponseFunc, logger lumber.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		errorHandler := transport.NewLogErrorHandler(logger)
		errorEncoder := gohttp.DefaultErrorEncoder
		request, err := dec(c, c, logger)
		if err != nil {
			errorHandler.Handle(c, err)
			errorEncoder(c, err, c.Writer)
			return
		}

		response, err := ep(c, request)
		if err != nil {
			errorHandler.Handle(c, err)
			errorEncoder(c, err, c.Writer)
			return
		}

		if err := enc(c, c, response, logger); err != nil {
			errorHandler.Handle(c, err)
			errorEncoder(c, err, c.Writer)
			return
		}
	}
}
