package main

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

type response map[string]interface{}

// TODO: Add tests in utils_http_test.go

func bodyUnmarshal(r *http.Request, x interface{}) error {
	return bodyUnmarshalOptionalFields(r, x, true, []string{})
}

func bodyUnmarshalOptional(r *http.Request, x interface{}) error {
	return bodyUnmarshalOptionalFields(r, x, false, []string{})
}

func bodyUnmarshalOptionalFields(r *http.Request, x interface{}, required bool, optional []string) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("cannot read POST body: %v\n", err)
		return errorInternal
	}

	if err = json.Unmarshal(b, x); err != nil {
		return errorInvalidJSONBody
	}

	// If fields are required, create a set of provided optional fields for quick lookup
	optionalFields := make(map[string]bool)
	if required {
		for _, field := range optional {
			optionalFields[field] = true
		}
	}

	xv := reflect.Indirect(reflect.ValueOf(x))
	xt := xv.Type()
	for i := 0; i < xv.NumField(); i++ {
		fieldName := xt.Field(i).Name
		if required && xv.Field(i).IsNil() && !optionalFields[fieldName] {
			return errorMissingField
		}
	}

	return nil
}

func bodyMarshal(w http.ResponseWriter, x map[string]interface{}) error {
	resp, err := json.Marshal(x)
	if err != nil {
		w.Write([]byte(`{"success":false,"message":"Some internal error occurred"}`))
		logger.Errorf("cannot marshal response: %v\n")
		return errorInternal
	}

	w.Write(resp)
	return nil
}

func getIp(r *http.Request) string {
	ip := r.RemoteAddr
	if r.Header.Get("X-Forwarded-For") != "" {
		ip = r.Header.Get("X-Forwarded-For")
	}

	return ip
}

func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}
