package main

import (
	"os"

	"github.com/x-formation/pulsekit/cli"
)

func main() {
	cli.New().Run(os.Args)
}
