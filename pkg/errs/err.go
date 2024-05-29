package errs

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Err is a custom error object complying to Headere and StatusCoder interfaces
// so that it can be readily used by custom HTTPHandler implementation
type Err struct {
	ID          string
	Code        int
	Message     string
	HTTPHeaders http.Header
}

func (e Err) Error() string {
	return fmt.Sprintf("%d : %s ", e.Code, e.Message)
}
func (e Err) Headers() http.Header {
	return e.HTTPHeaders
}

func (e Err) StatusCode() int {
	if e.Code > 0 {
		return e.Code
	} else {
		return 200
	}
}

func (e Err) MarshalJSON() ([]byte, error) {
	val := struct {
		ID      string
		Code    int
		Message string
	}{ID: e.ID, Code: e.Code, Message: e.Message}
	return json.Marshal(val)
}

var ERR_DUMMY = Err{
	ID:          "ERR::DUMMY",
	Message:     "Dummy error ",
	HTTPHeaders: http.Header{"x-some-header": []string{"exemplar"}},
	Code:        512}

var ERR_INVALID_ENVIRONMENT = Err{
	ID:      "ERR::INV::ENV",
	Message: "Invalid environment specified"}

var ERR_INF_API_MAX_ATTEMPT = Err{
	ID:      "ERR::INF::API::MAX::ATTEMPT",
	Message: "api server restart max attempt reached"}
