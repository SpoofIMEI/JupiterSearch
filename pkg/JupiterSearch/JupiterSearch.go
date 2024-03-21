package jupitersearch

import (
	"github.com/R00tendo/JupiterSearch/pkg/JupiterSearch/client"
)

func NewClient(server string, key string) (*client.Client, error) {
	Jclient := &client.Client{
		Server: server,
		Key:    key,
	}

	if err := Jclient.Check(); err != nil {
		return nil, err
	}

	return Jclient, nil
}
