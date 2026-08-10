package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// status is the response status. Conforms to JSend specification.
type status string

// Code is an error code. Can be created with [NewCode].
type Code *int

// JSend status values.
const (
	StatusSuccess status = "success"
	StatusFail    status = "fail"
	StatusError   status = "error"
)

// response is a JSend response,
type response struct {
	// Required always.
	Status string `json:"status"`

	// Required (even if null) for success and fail response, optional for error ones.
	Data *any `json:"data,omitempty"`

	// Required for error responses.
	Message *string `json:"message,omitempty"`

	// Optional for error responses.
	Code *int `json:"code,omitempty"`
}

// WriteSuccess writes a success response to w. If an error occurs encoding the response
// it will be returned.
func WriteSuccess(w http.ResponseWriter, data any) error {
	resp := response{
		Status: string(StatusSuccess),
		Data:   &data,
	}

	return writeResp(w, &resp)
}

// WriteFail writes a fail response to w. If an error occurs encoding the response it will
// be returned.
func WriteFail(w http.ResponseWriter, data any) error {
	resp := response{
		Status: string(StatusFail),
		Data:   &data,
	}

	return writeResp(w, &resp)
}

// WriteError writes an error response to w. If an error occurs encoding the response it will
// be returned.
//
// If code is set to nil it will be omitted, use [NewCode] to create new error codes.
//
//	func Handler(w http.ResponseWriter, r *http.Request) {
//		type Data struct {
//			Name string `json:"name"`
//	  }
//
//	  err := WriteError(w, "internal server error", NewCode(1), new(Data{Name: "arya"}))
//	  if err != nil {
//	    // log error...
//	  }
//	 }
func WriteError(w http.ResponseWriter, message string, code Code, data *any) error {
	resp := response{
		Status:  string(StatusError),
		Data:    data,
		Code:    (*int)(code),
		Message: &message,
	}

	return writeResp(w, &resp)
}

// NewCode creates a new [Code] from integer.
func NewCode(c int) Code {
	return Code(&c)
}

func writeResp(w http.ResponseWriter, resp *response) error {
	ContentJSON(w)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return fmt.Errorf("could not encode response: %w", err)
	}

	return nil
}
