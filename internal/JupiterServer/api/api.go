package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/R00tendo/JupiterSearch/internal/JupiterServer/index"
	"github.com/R00tendo/JupiterSearch/internal/JupiterServer/nodemaster"
	"github.com/R00tendo/JupiterSearch/internal/universal/information"
	"github.com/R00tendo/JupiterSearch/internal/universal/jhttp"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	serverStartTime time.Time
	APISrv          *http.Server
	ShutdownChan    = make(chan bool, 1)
	Shutdown        bool
)

func Start(wg *sync.WaitGroup) error {
	wg.Add(1)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(keyCheck())

	router.GET("/", root)
	router.GET("/shutdown", shutdown)
	router.GET("/nodes", nodes)
	router.POST("/store", store)
	router.POST("/search", search)

	APISrv = &http.Server{
		Addr:    viper.GetString("api_listen"),
		Handler: router,
	}

	serverStartTime = time.Now()

	var GinError error

	if viper.GetString("tls_cert") != "" {
		if viper.GetString("tls_private") == "" {
			return errors.New("tls_key not defined")
		}

		go func() {
			if err := APISrv.ListenAndServeTLS(viper.GetString("tls_cert"), viper.GetString("tls_private")); err != nil && err != http.ErrServerClosed {
				GinError = err
				wg.Done()
			}
		}()

	} else {
		go func() {
			if err := APISrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				GinError = err
				wg.Done()
			}
		}()
	}

	time.Sleep(500 * time.Millisecond)

	if GinError == nil {
		logrus.Info("API successfully started:", viper.GetString("api_listen"))
	}

	return GinError
}

func Stop() {
	if Shutdown {
		for _, node := range nodemaster.ConnectedNodes {
			jhttp.Request(node.ShutdownURL, nil)
		}
	}
	APISrv.Shutdown(context.Background())
}

func keyCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header["Jupiter-Key"] != nil {
			if c.Request.Header["Jupiter-Key"][0] == viper.GetString("client_key") {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, map[string]string{
			"error": "unauthorized!",
		})
	}
}

func shutdown(c *gin.Context) {
	c.JSON(200, map[string]string{
		"message": "Ok",
	})
	ShutdownChan <- true
}

func nodes(c *gin.Context) {
	c.JSON(200, nodemaster.ConnectedNodes)
}

var InfoPage = struct {
	Version string
	Nodes   int
	Uptime  string
}{
	information.ServerVersionNumber,
	0,
	"0m",
}

func root(c *gin.Context) {
	rootResp := InfoPage
	if int(time.Since(serverStartTime).Minutes()) < 60 {
		rootResp.Uptime = fmt.Sprintf(
			"%dmin",
			int(time.Since(serverStartTime).Minutes()),
		)

		rootResp.Nodes = len(nodemaster.ConnectedNodes)
	} else {
		rootResp.Uptime = fmt.Sprintf(
			"%dh",
			int(time.Since(serverStartTime).Hours()),
		)

		rootResp.Nodes = len(nodemaster.ConnectedNodes)
	}

	c.JSON(200, rootResp)
}

func store(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Debug(err.Error())
		c.JSON(200, map[string]string{
			"error": err.Error(),
		})
		return
	}

	request := make(map[string]any)
	err = json.Unmarshal(data, &request)
	if err != nil {
		logrus.Debug(err.Error())
		c.JSON(200, map[string]string{
			"error": err.Error(),
		})

		return
	}

	ID, err := index.Index(request)
	if err != nil {
		logrus.Debug(err.Error())
		c.JSON(200, map[string]string{
			"error": err.Error(),
		})

		return
	}

	c.JSON(200, map[string]string{
		"message": ID,
	})
}

func search(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Debug(err.Error())

		c.JSON(200, map[string]string{
			"error": err.Error(),
		})

		return
	}

	request := make(map[string]any)
	err = json.Unmarshal(data, &request)
	if err != nil {
		logrus.Debug(err.Error())

		c.JSON(200, map[string]string{
			"error": err.Error(),
		})

		return
	}

	logrus.Debug("new search query:", request["query"].(string))

	if len(nodemaster.ConnectedNodes) == 0 {
		c.JSON(200, map[string]string{
			"error": "no nodes connected",
		})

		return
	}

	queryMessage := map[string]any{
		"command": "query",
		"query":   request["query"].(string),
	}

	results := make(map[string]any)
	var wg sync.WaitGroup

	for _, node := range nodemaster.ConnectedNodes {
		resp, err := jhttp.Request(node.SearchURL, queryMessage)
		if err != nil {
			nodemaster.Remove(node.ID)
			continue
		}

		if resp["error"] != nil {
			results[node.Name] = "error:" + resp["error"].(string)
			continue
		}

		results[node.Name] = resp["results"].(string)
	}

	wg.Wait()

	c.JSON(200, map[string]any{
		"results": results,
	})
}
