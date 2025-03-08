package client

import (
	"errors"
	"net/url"

	"github.com/SpoofIMEI/JupiterSearch/pkg/JupiterSearch/client/httpcli"
)

type Client struct {
	Server string
	Key    string
}

func (Jclient *Client) request(Path string, data map[string]any) (map[string]any, error) {
	URL, _ := url.JoinPath(Jclient.Server, Path)

	if len(data) == 0 {
		return httpcli.GetRequest(URL, Jclient.Key)
	} else {
		return httpcli.PostReguest(URL, data, Jclient.Key)
	}
}

func (Jclient *Client) Check() error {
	resp, err := Jclient.request("", nil)
	if err != nil {
		return err
	}

	if resp["error"] != nil {
		return errors.New(resp["error"].(string))
	}

	return nil
}

func (Jclient *Client) Search(query string) (map[string]any, error) {
	resp, err := Jclient.request("search", map[string]any{
		"query": query,
	})
	if err != nil {
		return nil, err
	}

	if resp["results"] == nil {
		if resp["error"] != nil {
			return nil, errors.New(resp["error"].(string))
		}
		return nil, errors.New("invalid response")
	}

	return resp["results"].(map[string]any), nil
}

func (Jclient *Client) Store(data map[string]any) (string, error) {
	resp, err := Jclient.request("store", map[string]any{
		"store": data,
	})
	if err != nil {
		return "", err
	}

	if resp["message"] != nil {
		return resp["message"].(string), nil
	} else if resp["error"] != nil {
		return "", errors.New(resp["error"].(string))
	} else {
		return "", errors.New("server sent an invalid response")
	}
}

func (Jclient *Client) Shutdown() error {
	_, err := Jclient.request("shutdown", nil)
	if err != nil {
		return err
	}
	return nil
}
