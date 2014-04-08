package main

import (
	"os"

	"github.com/x-formation/int-tools/pulseutil/pulsecli"
)

func main() {
	pulsecli.New().Run(os.Args)
}
