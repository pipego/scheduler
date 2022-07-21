package main

import (
	"fmt"
	"os"

	"github.com/pipego/scheduler/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
