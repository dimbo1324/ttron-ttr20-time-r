package main

import (
	"os"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/app/clientapp"
)

func main() {
	os.Exit(clientapp.Run(os.Args[1:]))
}
