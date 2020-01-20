package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func decode(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

func encode(w io.Writer, v interface{}) error {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return err
	}
	return nil
}

// JSONError represents an error in JSON format.
type JSONError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func jsonError(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	err := &JSONError{Code: code}
	strErr, ok := v.(string)
	if ok {
		err.Msg = strErr
		_ = encode(w, err)
		return
	}
}
