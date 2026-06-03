package logging

import (
	"log"

	platformlogging "github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/logging"
)

func New(path string) *log.Logger {
	return platformlogging.New(path)
}
