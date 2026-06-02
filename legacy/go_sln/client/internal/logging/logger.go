package logging

import (
	"log"
	"os"
	"path/filepath"
)

// New возвращает *log.Logger
// path == "" -> лог в stdout
// Если path - только имя файла -> создаём <project-root>/logs/<name>
// Если path содержит путь -> используем его (создаём папки)
func New(path string) *log.Logger {
	// пустой путь - stdout
	if path == "" {
		return log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	}

	// если path - только имя файла (без разделителей)
	if filepath.Base(path) == path {
		// пытаемся получить текущую директорию
		cwd, err := os.Getwd()
		if err != nil {
			// при ошибке - попытка открыть файл в cwd
			fallback, ferr := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if ferr != nil {
				// если и это не удалось - лог в stdout
				l := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
				l.Printf("cannot open log file %s: %v (cwd error: %v) - logging to stdout", path, ferr, err)
				return l
			}
			return log.New(fallback, "", log.LstdFlags|log.Lmicroseconds)
		}

		// если cwd - server или client, считаем корнем родительскую папку
		base := filepath.Base(cwd)
		projectRoot := cwd
		if base == "server" || base == "client" {
			projectRoot = filepath.Dir(cwd)
		}

		logsDir := filepath.Join(projectRoot, "logs")
		// создаём logs/, если нужно
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			l := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
			l.Printf("cannot create logs dir %s: %v - logging to stdout", logsDir, err)
			return l
		}

		filePath := filepath.Join(logsDir, path)
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			l := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
			l.Printf("cannot open log file %s: %v - logging to stdout", filePath, err)
			return l
		}
		return log.New(f, "", log.LstdFlags|log.Lmicroseconds)
	}

	// path содержит директорию - создаём родительские папки и открываем файл
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		l := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
		l.Printf("cannot create log directory %s: %v - logging to stdout", dir, err)
		return l
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		l := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
		l.Printf("cannot open log file %s: %v - logging to stdout", path, err)
		return l
	}
	return log.New(f, "", log.LstdFlags|log.Lmicroseconds)
}
