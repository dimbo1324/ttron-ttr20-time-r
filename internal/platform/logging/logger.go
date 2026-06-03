package logging

import (
	"log"
	"os"
	"path/filepath"
)

type Logger interface {
	Printf(format string, v ...any)
}

func New(path string) *log.Logger {
	if path == "" {
		return log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	}

	if filepath.Base(path) == path {
		cwd, err := os.Getwd()
		if err != nil {
			return stdoutWithNotice("cannot resolve cwd for log file %s: %v", path, err)
		}
		path = filepath.Join(cwd, "logs", path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return stdoutWithNotice("cannot create log directory for %s: %v", path, err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return stdoutWithNotice("cannot open log file %s: %v", path, err)
	}
	return log.New(f, "", log.LstdFlags|log.Lmicroseconds)
}

func stdoutWithNotice(format string, args ...any) *log.Logger {
	l := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	l.Printf(format+" - logging to stdout", args...)
	return l
}
