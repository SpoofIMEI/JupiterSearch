package index

import (
	"errors"
	"reflect"

	"github.com/R00tendo/JupiterSearch/internal/JupiterServer/nodemaster"
	"github.com/R00tendo/JupiterSearch/internal/universal/jhttp"
)

func Index(data map[string]any) (string, error) {
	var smallestSize float64
	var smallestNode *nodemaster.Node

	for _, node := range nodemaster.ConnectedNodes {
		if smallestSize == 0 {
			smallestSize, smallestNode = node.DBsize, node
			continue
		}

		if node.DBsize < smallestSize {
			smallestSize, smallestNode = node.DBsize, node
		}
	}

	if smallestNode != nil {
		var resp map[string]any

		v := reflect.TypeOf(data["store"])
		switch v.Kind() {

		case reflect.Map:
			var err error

			resp, err = jhttp.Request(smallestNode.StoreURL, map[string]any{
				"store": data["store"].(map[string]any),
			})
			if err != nil {
				return "", err
			}

		default:
			return "", errors.New("invalid JSON")
		}

		switch ID := resp["message"].(type) {
		case string:
			return ID, nil
		}

		if resp["error"] != nil {
			return "", errors.New(resp["error"].(string))
		}

		return "", errors.New("server sent invalid response")
	}

	return "", errors.New("no nodes connected")
}
