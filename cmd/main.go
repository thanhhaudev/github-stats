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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := container.NewDataContainer()
	err := d.Build(ctx)
	if err != nil {
		panic(err)
	}

	err = writer.UpdateReadme(d.GetStats(), os.Getenv("SECTION_NAME"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Execution Duration: %s\n", time.Since(start))
}
