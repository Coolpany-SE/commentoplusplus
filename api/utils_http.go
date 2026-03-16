package main

import (
	"io"
	"net/http"
)

type response map[string]interface{}

func bodyRead(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("cannot read POST body: %v\n", err)
		return nil, errorInternal
	}

	return b, nil
}

func bodyMarshal(w http.ResponseWriter, x map[string]interface{}) error {
	resp, err := MarshalJson(x)
	w.Write(resp)

	return err
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

// Deprecated

func bodyUnmarshal(r *http.Request, x interface{}) error {
	return bodyUnmarshalRequest(r, x, true, []string{})
}

func bodyUnmarshalOptional(r *http.Request, x interface{}) error {
	return bodyUnmarshalRequest(r, x, false, []string{})
}

func bodyUnmarshalRequest(r *http.Request, x interface{}, required bool, optional []string) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("cannot read POST body: %v\n", err)
		return errorInternal
	}

	return UnmarshalAndValidateJson(b, x, required, optional)
}
