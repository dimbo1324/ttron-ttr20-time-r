package config

import (
	"flag"
	"fmt"
	"os"
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
	GRPCListen     string
}

func DefaultGateway() GatewayConfig {
	return GatewayConfig{
		Target:         "127.0.0.1:9000",
		CRCMode:        "sum",
		AdapterAddr:    1,
		PollInterval:   5 * time.Second,
		RequestTimeout: 1500 * time.Millisecond,
		ConnectTimeout: 2 * time.Second,
		BackoffInitial: 500 * time.Millisecond,
		BackoffMax:     5 * time.Second,
		RecentSize:     100,
		LogFile:        "runtime/logs/ft12-gateway.log",
		GRPCListen:     ":9200",
	}
}

func LoadGateway(args []string) (*GatewayConfig, error) {
	c := DefaultGateway()
	fs := flag.NewFlagSet("ft12-gateway", flag.ContinueOnError)
	fs.StringVar(&c.Target, "target", c.Target, "target emulator/device TCP address")
	fs.StringVar(&c.CRCMode, "crc", c.CRCMode, "crc mode: sum | crc16")
	mode := fs.String("mode", "", "checksum mode alias: sum | crc16")
	fs.IntVar(&c.AdapterAddr, "adapter", c.AdapterAddr, "adapter address byte (0..255)")
	fs.DurationVar(&c.PollInterval, "interval", c.PollInterval, "polling interval")
	fs.DurationVar(&c.RequestTimeout, "timeout", c.RequestTimeout, "request/response timeout")
	fs.DurationVar(&c.ConnectTimeout, "connect-timeout", c.ConnectTimeout, "TCP connect timeout")
	fs.DurationVar(&c.BackoffInitial, "backoff-initial", c.BackoffInitial, "initial reconnect backoff")
	fs.DurationVar(&c.BackoffMax, "backoff-max", c.BackoffMax, "maximum reconnect backoff")
	fs.IntVar(&c.RecentSize, "recent", c.RecentSize, "recent frame/event buffer size")
	fs.StringVar(&c.LogFile, "log", c.LogFile, "path to log file; empty = stdout")
	fs.StringVar(&c.GRPCListen, "grpc-listen", c.GRPCListen, "gRPC control listen address; empty disables gRPC")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *mode != "" {
		c.CRCMode = *mode
	}
	c.Normalize()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func MustLoadGatewayFromOS() *GatewayConfig {
	cfg, err := LoadGateway(os.Args[1:])
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *GatewayConfig) Normalize() {
	if c.CRCMode == "" {
		c.CRCMode = "sum"
	}
}

func (c GatewayConfig) Validate() error {
	if c.Target == "" {
		return fmt.Errorf("target must not be empty")
	}
	if err := validateTCPAddress(c.Target, "target address"); err != nil {
		return err
	}
	if err := validateChecksumMode(c.CRCMode); err != nil {
		return err
	}
	if err := validateAdapterAddr(c.AdapterAddr); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.PollInterval, "poll interval"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.RequestTimeout, "request timeout"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.ConnectTimeout, "connect timeout"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.BackoffInitial, "backoff initial"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.BackoffMax, "backoff max"); err != nil {
		return err
	}
	if c.BackoffInitial > c.BackoffMax {
		return fmt.Errorf("backoff initial must not exceed backoff max")
	}
	if err := validateRecentSize(c.RecentSize); err != nil {
		return err
	}
	if c.GRPCListen != "" {
		if err := validateTCPAddress(c.GRPCListen, "gRPC listen address"); err != nil {
			return err
		}
	}
	return nil
}
