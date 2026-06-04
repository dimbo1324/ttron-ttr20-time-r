package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	url := flag.String("url", "http://127.0.0.1:8080/health", "HTTP URL to check")
	tcp := flag.String("tcp", "", "TCP host:port to check instead of HTTP")
	timeout := flag.Duration("timeout", 2*time.Second, "request timeout")
	flag.Parse()

	if *tcp != "" {
		conn, err := net.DialTimeout("tcp", *tcp, *timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tcp healthcheck failed: %v\n", err)
			os.Exit(1)
		}
		_ = conn.Close()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck request failed: %v\n", err)
		os.Exit(1)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "healthcheck returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}
}
