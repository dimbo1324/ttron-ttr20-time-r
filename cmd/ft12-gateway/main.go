package main

import (
	"os"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/app/gatewayapp"
)

func main() {
	os.Exit(gatewayapp.Run(os.Args[1:]))
}
