package sagas

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	ErrInvalidHttpMethod = errors.New("invalid HTTP method")
)

func HttpReq(url, method, requestID string, body map[string]string) error {
	client := http.Client{}

	switch strings.ToUpper(method) {
	case "GET":
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("request-id", requestID)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Println(respBody)
	case "post":
		reqBody, err := json.Marshal(body)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("request-id", requestID)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Println(respBody)
	default:
		return ErrInvalidHttpMethod
	}

	return nil
}
