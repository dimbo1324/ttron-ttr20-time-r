package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCreatesDirectoryAndWritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "service.log")

	logger := New(path)
	logger.Print("hello log")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "hello log") {
		t.Fatalf("log file does not contain message: %q", string(data))
	}
}

func TestNewUsesRuntimeLogsForBareFilename(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	logger := New("service.log")
	logger.Print("runtime message")

	path := filepath.Join(dir, "runtime", "logs", "service.log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime log file: %v", err)
	}
	if !strings.Contains(string(data), "runtime message") {
		t.Fatalf("runtime log file does not contain message: %q", string(data))
	}
}

func TestNewWithEmptyPathReturnsStdoutLogger(t *testing.T) {
	logger := New("")
	if logger == nil {
		t.Fatal("expected stdout logger")
	}
	logger.Print("stdout fallback is usable")
}

func TestNewReturnsFallbackWhenFileCannotOpen(t *testing.T) {
	dir := t.TempDir()
	blockingFile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	logger := New(filepath.Join(blockingFile, "service.log"))
	if logger == nil {
		t.Fatal("expected fallback logger")
	}
	logger.Print("fallback is usable")
}

func TestDefaultFile(t *testing.T) {
	got := DefaultFile("ft12-api")
	want := filepath.Join("runtime", "logs", "ft12-api.log")
	if got != want {
		t.Fatalf("DefaultFile() = %q, want %q", got, want)
	}
}
