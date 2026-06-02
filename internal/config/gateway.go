package config

import (
	"flag"
	"time"
)

type GatewayConfig struct {
	Target         string
	CRCMode        string
	AdapterAddr    int
	PollInterval   time.Duration
	RequestTimeout time.Duration
	ConnectTimeout time.Duration
	BackoffInitial time.Duration
	BackoffMax     time.Duration
	RecentSize     int
	LogFile        string
}

func LoadGateway() *GatewayConfig {
	c := &GatewayConfig{}
	flag.StringVar(&c.Target, "target", "127.0.0.1:9000", "target emulator/device TCP address")
	flag.StringVar(&c.CRCMode, "crc", "sum", "crc mode: sum | crc16")
	mode := flag.String("mode", "", "checksum mode alias: sum | crc16")
	flag.IntVar(&c.AdapterAddr, "adapter", 1, "adapter address byte (0..255)")
	flag.DurationVar(&c.PollInterval, "interval", 5*time.Second, "polling interval")
	flag.DurationVar(&c.RequestTimeout, "timeout", 1500*time.Millisecond, "request/response timeout")
	flag.DurationVar(&c.ConnectTimeout, "connect-timeout", 2*time.Second, "TCP connect timeout")
	flag.DurationVar(&c.BackoffInitial, "backoff-initial", 500*time.Millisecond, "initial reconnect backoff")
	flag.DurationVar(&c.BackoffMax, "backoff-max", 5*time.Second, "maximum reconnect backoff")
	flag.IntVar(&c.RecentSize, "recent", 100, "recent frame/event buffer size")
	flag.StringVar(&c.LogFile, "log", "", "path to log file; empty = stdout")
	flag.Parse()
	if *mode != "" {
		c.CRCMode = *mode
	}
	return c
}
