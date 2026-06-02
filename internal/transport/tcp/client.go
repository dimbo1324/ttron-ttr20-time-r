package tcp

import (
	"context"
	"net"
	"time"
)

type ClientConfig struct {
	Address        string
	ConnectTimeout time.Duration
}

func Dial(ctx context.Context, cfg ClientConfig) (net.Conn, error) {
	timeout := cfg.ConnectTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	dialer := net.Dialer{Timeout: timeout}
	return dialer.DialContext(ctx, "tcp", cfg.Address)
}
