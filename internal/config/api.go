package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"
)

type APIConfig struct {
	HTTPListen     string
	EmulatorGRPC   string
	GatewayGRPC    string
	RequestTimeout time.Duration
	CORSOrigin     string
	LogFile        string
}

func DefaultAPI() APIConfig {
	return APIConfig{
		HTTPListen:     ":8080",
		EmulatorGRPC:   "127.0.0.1:9100",
		GatewayGRPC:    "127.0.0.1:9200",
		RequestTimeout: 3 * time.Second,
		CORSOrigin:     "http://localhost:5173",
		LogFile:        "runtime/logs/ft12-api.log",
	}
}

func LoadAPI(args []string) (*APIConfig, error) {
	c := DefaultAPI()
	fs := flag.NewFlagSet("ft12-api", flag.ContinueOnError)
	fs.StringVar(&c.HTTPListen, "http-listen", c.HTTPListen, "HTTP API listen address")
	fs.StringVar(&c.EmulatorGRPC, "emulator-grpc", c.EmulatorGRPC, "emulator gRPC address")
	fs.StringVar(&c.GatewayGRPC, "gateway-grpc", c.GatewayGRPC, "gateway gRPC address")
	fs.DurationVar(&c.RequestTimeout, "timeout", c.RequestTimeout, "per-request upstream timeout")
	fs.StringVar(&c.CORSOrigin, "cors-origin", c.CORSOrigin, "allowed CORS origin, empty disables CORS")
	fs.StringVar(&c.LogFile, "log", c.LogFile, "path to log file; empty = stdout")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.Normalize()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func MustLoadAPIFromOS() *APIConfig {
	cfg, err := LoadAPI(os.Args[1:])
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *APIConfig) Normalize() {}

func (c APIConfig) Validate() error {
	if c.HTTPListen == "" {
		return fmt.Errorf("http listen address must not be empty")
	}
	if err := validateTCPAddress(c.HTTPListen, "http listen address"); err != nil {
		return err
	}
	if c.EmulatorGRPC == "" {
		return fmt.Errorf("emulator gRPC address must not be empty")
	}
	if err := validateTCPAddress(c.EmulatorGRPC, "emulator gRPC address"); err != nil {
		return err
	}
	if c.GatewayGRPC == "" {
		return fmt.Errorf("gateway gRPC address must not be empty")
	}
	if err := validateTCPAddress(c.GatewayGRPC, "gateway gRPC address"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.RequestTimeout, "request timeout"); err != nil {
		return err
	}
	return validateOrigin(c.CORSOrigin)
}

func validateOrigin(origin string) error {
	if origin == "" || origin == "*" {
		return nil
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" || u.Path != "" {
		return fmt.Errorf("cors origin must be empty, *, or a valid origin")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("cors origin scheme must be http or https")
	}
	return nil
}
