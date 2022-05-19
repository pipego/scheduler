package main

import (
	"context"
	"os"

	"github.com/pipego/scheduler/cmd"
)

func main() {
	ctx := context.Background()

	if err := cmd.Run(ctx); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
