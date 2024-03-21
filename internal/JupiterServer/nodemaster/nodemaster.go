package nodemaster

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/R00tendo/JupiterSearch/internal/universal/jhttp"
	"github.com/R00tendo/JupiterSearch/internal/universal/keys"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Node struct {
	Name        string
	DBsize      float64
	ID          string
	RootURL     string
	StoreURL    string
	SearchURL   string
	ShutdownURL string
}

var (
	ConnectedNodes = make(map[string]*Node)
	BlackList      = make(map[string]int)
	Retry          bool
)

func Remove(ID string) {
	logrus.Info("removing node:", ConnectedNodes[ID].Name)

	BlackList[ConnectedNodes[ID].RootURL] = 20

	delete(ConnectedNodes, ID)
}

func CheckNodes() {
	nodesConnected := make(map[string]string)

	for _, node := range viper.GetStringSlice("nodes") {
		for _, connectedNode := range ConnectedNodes {
			if node == connectedNode.RootURL {
				nodesConnected[node] = connectedNode.ID
			}
		}
	}

	var atleastOneChecked bool
	for _, node := range viper.GetStringSlice("nodes") {
		if BlackList[node] > 10 {
			continue
		}

		atleastOneChecked = true

		connected := len(nodesConnected[node]) > 0

		info, err := jhttp.Request(node, nil)
		if err != nil {
			if !Retry {
				BlackList[node]++
				if BlackList[node] > 10 {
					logrus.Info("giving up on ", node)
					continue
				}

				logrus.Error("node:", node, " doesn't answer")
				if connected {
					Remove(nodesConnected[node])
					continue
				}
			} else {
				logrus.Debug("node:", node, " doesn't answer")
			}

			continue
		}

		if !keys.Contains(info, []string{"name", "dbsize"}) {
			if keys.Contains(info, []string{"error"}) {
				logrus.Debug(info["error"].(string))
				continue
			}

			if !Retry {
				BlackList[node]++
				if BlackList[node] > 10 {
					logrus.Info("giving up on ", node)
					continue
				}

				logrus.Error("node:", node, " sent invalid info, won't count it as alive")

				if connected && !Retry {
					Remove(nodesConnected[node])
				}
			} else {
				logrus.Debug("node:", node, " sent invalid info, won't count it as alive")
			}

			continue
		}

		var ID string

		if !connected {
			ID = uuid.New().String()
		} else {
			ID = ConnectedNodes[nodesConnected[node]].ID
		}

		StoreURL, _ := url.JoinPath(node, "store")
		SearchURL, _ := url.JoinPath(node, "search")
		ShutdownURL, _ := url.JoinPath(node, "shutdown")
		Name := info["name"].(string)

		if !connected {
			var suffix int
			tempName := Name

			for {
				var taken bool

				for _, node := range ConnectedNodes {
					if node.Name == tempName {
						suffix++
						tempName = fmt.Sprintf("%s_%d", Name, suffix)
						taken = true
					}
				}

				if !taken {
					Name = tempName
					break
				}
			}
		}

		nodeObj := &Node{
			Name:        Name,
			DBsize:      info["dbsize"].(float64),
			ShutdownURL: ShutdownURL,
			StoreURL:    StoreURL,
			SearchURL:   SearchURL,
			RootURL:     node,
			ID:          ID,
		}

		ConnectedNodes[ID] = nodeObj
		if !connected {
			logrus.Info(node, " is alive!")
		}
	}
	if !atleastOneChecked && !Retry {
		logrus.Error("no nodes available")
		os.Exit(0)
	}
}

func Checker() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			CheckNodes()
		}
	}()
}
