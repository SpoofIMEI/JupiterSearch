package jhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"

	"github.com/spf13/viper"
)

func Request(URL string, data map[string]any) (map[string]any, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if len(data) == 0 {
		return getRequest(URL)
	} else {
		return postReguest(URL, data)
	}
}

func postReguest(URL string, body map[string]any) (map[string]any, error) {
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyJSON := bytes.NewBuffer(bodyRaw)

	req, err := http.NewRequest("POST", URL, bodyJSON)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Jupiter-Key", viper.GetString("node_key"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respJSON map[string]any

	err = json.Unmarshal(respBytes, &respJSON)
	if err != nil {
		return nil, err
	}

	return respJSON, nil
}

func getRequest(URL string) (map[string]any, error) {
	req, err := http.NewRequest("GET", URL, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Jupiter-Key", viper.GetString("node_key"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	bytesJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respJSON map[string]any
	err = json.Unmarshal(bytesJSON, &respJSON)
	if err != nil {
		return nil, err
	}

	return respJSON, nil
}
