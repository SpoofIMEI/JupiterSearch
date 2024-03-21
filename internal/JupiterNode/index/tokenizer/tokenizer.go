package tokenizer

import (
	"regexp"
	"strings"
)

var Regex *regexp.Regexp

func Tokenize(text string) []string {
	var results []string
	matches := make(map[string]bool)

	for _, match := range Regex.FindAllString(text, -1) {
		match = strings.TrimSpace(match)
		matches[match] = true
	}

	for match := range matches {
		results = append(results, match)
	}

	return results
}
