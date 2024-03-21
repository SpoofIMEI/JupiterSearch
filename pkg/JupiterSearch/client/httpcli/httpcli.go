package httpcli

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
)

func PostReguest(URL string, body map[string]any, Key string) (map[string]any, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyJSON := bytes.NewBuffer(bodyRaw)

	req, err := http.NewRequest("POST", URL, bodyJSON)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Jupiter-Key", Key)

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

func GetRequest(URL string, Key string) (map[string]any, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", URL, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Jupiter-Key", Key)

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
