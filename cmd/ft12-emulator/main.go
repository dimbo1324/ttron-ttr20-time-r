package main

import (
	"os"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/app/emulatorapp"
)

func main() {
	os.Exit(emulatorapp.Run(os.Args[1:]))
}
