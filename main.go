package main

import (
	"os"

	"github.com/pipego/scheduler/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
