package main

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/clock"
	"github.com/thanhhaudev/github-stats/pkg/container"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

func main() {
	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl := clock.NewClock()
	if tz := os.Getenv("TIME_ZONE"); tz != "" {
		err := cl.SetLocation(tz)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Timezone set to %s\n", tz)
	}

	d := container.NewDataContainer()
	if err := d.Build(ctx); err != nil {
		panic(err)
	}

	err := writer.UpdateReadme(d.GetStats(cl), os.Getenv("SECTION_NAME"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Execution Duration: %s\n", time.Since(start))
}
