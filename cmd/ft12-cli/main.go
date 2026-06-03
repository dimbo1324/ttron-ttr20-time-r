package main

import (
	"os"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/app/cliapp"
)

func main() {
	os.Exit(cliapp.Run(os.Args[1:]))
}
