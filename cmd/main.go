package main

import (
	"context"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/container"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

func main() {
	ctx := context.Background()
	d := container.NewDataContainer(ctx)
	d.Build()

	err := writer.UpdateReadme(d.GetStats(), os.Getenv("SECTION_NAME"))
	if err != nil {
		panic(err)
	}
}
