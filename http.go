package sagas

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Errors for HTTP Requests
var (
	ErrInvalidHTTPMethod = errors.New("invalid HTTP method")
)

// HTTPReq issues an HTTP request based on the provided input
func HTTPReq(url, method, requestID string, body map[string]string) (map[string]string, error) {
	client := http.Client{}

	switch strings.ToUpper(method) {
	case "GET":
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("request-id", requestID)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		var result map[string]string

		json.NewDecoder(resp.Body).Decode(&result)

		return result, nil
	case "POST":
		reqBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("request-id", requestID)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		var result map[string]string

		json.NewDecoder(resp.Body).Decode(&result)

		return result, nil
	default:
		return nil, ErrInvalidHTTPMethod
	}
}
