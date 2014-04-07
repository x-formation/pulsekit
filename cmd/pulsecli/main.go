package main

import (
	"os"

	"github.com/x-formation/int-tools/pulse/pulsecli"
)

func main() {
	pulsecli.New().Run(os.Args)
}
