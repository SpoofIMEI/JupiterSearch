package config

import (
	"errors"
	"os"
	"regexp"

	"github.com/R00tendo/JupiterSearch/internal/JupiterNode/index/tokenizer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	configLocations = map[string][]string{
		"server": {
			"./JupiterServer.conf",
			"/etc/JupiterSearch/JupiterServer.conf",
		},
		"node": {
			"./JupiterNode.conf",
			"/etc/JupiterSearch/JupiterNode.conf",
		},
	}

	requiredParams = map[string][]string{
		"server": {
			"api_listen",
			"nodes",
			"node_key",
			"client_key",
		},
		"node": {
			"name",
			"datadir",
			"api_listen",
			"node_key",
			"max_concurrent_ingests",
		},
	}
)

func Parse(customConfigFilename string, mode string) error {
	var configFilesToCheck []string

	if customConfigFilename != "" {
		configFilesToCheck = append(configFilesToCheck, customConfigFilename)
	} else {
		configFilesToCheck = configLocations[mode]
	}

	var configFile string

	for _, _configFile := range configFilesToCheck {
		info, err := os.Stat(_configFile)
		if err != nil {
			continue
		} else if info.IsDir() {
			continue
		}

		configFile = _configFile
		logrus.Debug("config file found:", configFile)

		break
	}
	if configFile == "" {
		return errors.New("could not read configuration file")
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType("env")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	for _, requiredParam := range requiredParams[mode] {
		if viper.Get(requiredParam) == nil {
			return errors.New(requiredParam + " is not defined in configuration file")
		}
	}

	if mode == "node" {
		regexBytes, err := os.ReadFile("/etc/JupiterSearch/tokenization_regex")
		if err != nil {
			return err
		}

		tokenizer.Regex, err = regexp.Compile(string(regexBytes))
		if err != nil {
			return err
		}
	}

	return nil
}
