package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/writer"
	"os"
)

func main() {
	d := NewDataContainer()
	d.Build()

	err := writer.UpdateReadme(d.GetStats(), os.Getenv("SECTION_NAME"))
	if err != nil {
		panic(err)
	}
}
