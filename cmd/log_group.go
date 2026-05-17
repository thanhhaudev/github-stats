package main

import (
	"log"
	"strings"
)

func runGroupedStep(logger *log.Logger, title string, enabled bool, fn func() error) (err error) {
	if !enabled {
		return fn()
	}

	logger.Printf("::group::%s\n", escapeGroupTitle(title))
	defer logger.Println("::endgroup::")

	return fn()
}

func escapeGroupTitle(title string) string {
	replacer := strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
	)

	return replacer.Replace(title)
}
