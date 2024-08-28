package main

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/container"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

func main() {
	start := time.Now()

	ctx := context.Background()
	d := container.NewDataContainer(ctx)
	err := d.Build()
	if err != nil {
		panic(err)
	}

	err = writer.UpdateReadme(d.GetStats(), os.Getenv("SECTION_NAME"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Execution Duration: %s\n", time.Since(start))
}
