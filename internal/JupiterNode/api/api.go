package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/R00tendo/JupiterSearch/internal/JupiterNode/index"
	"github.com/R00tendo/JupiterSearch/internal/JupiterNode/index/database"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	APISrv   *http.Server
	Shutdown = make(chan bool, 1)
)

func Start(wg *sync.WaitGroup) error {
	wg.Add(1)

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(keyCheck())

	router.GET("/shutdown", shutdown)
	router.GET("/", root)
	router.POST("/store", store)
	router.POST("/search", search)

	APISrv = &http.Server{
		Addr:    viper.GetString("api_listen"),
		Handler: router,
	}

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
	if APISrv != nil {
		APISrv.Shutdown(context.Background())
	}
}

func keyCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header["Jupiter-Key"] != nil {
			if c.Request.Header["Jupiter-Key"][0] == viper.GetString("node_key") {
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
	Shutdown <- true
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

	switch query := request["query"].(type) {
	case string:
		logrus.Debug("new search query:", query)
		results, err := database.Retrieve(query)
		if err != nil {
			c.JSON(200, map[string]string{
				"error": err.Error(),
			})
			return
		}
		c.JSON(200, map[string]string{
			"results": results,
		})

	default:
		c.JSON(200, map[string]string{
			"error": "invalid JSON",
		})
	}
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

	v := reflect.TypeOf(request["store"])
	switch v.Kind() {
	case reflect.Map:
		ID, err := index.Index(request["store"].(map[string]any))
		if err != nil {
			c.JSON(200, map[string]string{
				"error": err.Error(),
			})

			return
		}

		c.JSON(200, map[string]string{
			"message": ID,
		})

	default:
		c.JSON(200, map[string]string{
			"error": "invalid JSON",
		})
	}
}

func root(c *gin.Context) {
	logrus.Debug("information request from master server:", c.RemoteIP())

	DBSize, err := database.DirSize()
	if err != nil {
		c.JSON(200, map[string]string{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, map[string]any{
		"name":   viper.GetString("name"),
		"dbsize": DBSize,
	})
}
