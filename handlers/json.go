package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra interface{}

	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON object")
		}

		return err
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, value interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(value)
}
