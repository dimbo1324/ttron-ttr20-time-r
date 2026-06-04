package main

import (
	"os"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/app/apiapp"
)

func main() {
	os.Exit(apiapp.Run(os.Args[1:]))
}
