package index

import (
	"errors"
	"strings"

	"github.com/R00tendo/JupiterSearch/internal/JupiterNode/index/database"
	"github.com/R00tendo/JupiterSearch/internal/JupiterNode/index/tokenizer"

	"github.com/sirupsen/logrus"
)

func Index(data map[string]any) (string, error) {
	lowercase := make(map[string]string)

	for name, text := range data {
		switch text.(type) {

		case string:
			lowercase[strings.ToLower(name)] = strings.ToLower(text.(string))

		default:
			return "", errors.New("invalid JSON document")
		}
	}

	tokenized := make(map[string][]string)
	for name, text := range lowercase {
		tokenized[name] = tokenizer.Tokenize(text)
	}

	ID, err := database.Store(tokenized, data)
	if err != nil {
		logrus.Error(err.Error())
	}
	return ID, nil
}
