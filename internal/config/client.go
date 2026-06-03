package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

type ClientConfig struct {
	Host         string
	Port         int
	CRCMode      string
	AdapterAddr  int
	TimeoutMs    int
	Retries      int
	LogFile      string
	PollEverySec int
}

func DefaultClient() ClientConfig {
	return ClientConfig{
		Host:         "127.0.0.1",
		Port:         9000,
		CRCMode:      "sum",
		AdapterAddr:  1,
		TimeoutMs:    1000,
		Retries:      2,
		PollEverySec: 1,
	}
}

func LoadClient(args []string) (*ClientConfig, error) {
	c := DefaultClient()
	fs := flag.NewFlagSet("ft12-client", flag.ContinueOnError)
	fs.StringVar(&c.Host, "host", c.Host, "server host")
	fs.IntVar(&c.Port, "port", c.Port, "server port")
	fs.StringVar(&c.CRCMode, "crc", c.CRCMode, "crc mode: sum | crc16")
	fs.IntVar(&c.AdapterAddr, "adapter", c.AdapterAddr, "adapter address (0..255)")
	fs.IntVar(&c.TimeoutMs, "timeout", c.TimeoutMs, "timeout for response in milliseconds")
	fs.IntVar(&c.Retries, "retries", c.Retries, "number of retries on timeout/error")
	fs.StringVar(&c.LogFile, "log", c.LogFile, "log file (empty = stdout)")
	fs.IntVar(&c.PollEverySec, "pollstep", c.PollEverySec, "polling tick step in seconds")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.Normalize()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func MustLoadClientFromOS() *ClientConfig {
	cfg, err := LoadClient(os.Args[1:])
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *ClientConfig) Normalize() {
	if c.CRCMode == "" {
		c.CRCMode = "sum"
	}
}

func (c ClientConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host must not be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be in range 1..65535")
	}
	if c.AdapterAddr < 0 || c.AdapterAddr > 255 {
		return fmt.Errorf("adapter address must be in range 0..255")
	}
	if _, err := checksum.ParseMode(c.CRCMode); err != nil {
		return err
	}
	if c.TimeoutMs <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.Retries < 0 {
		return fmt.Errorf("retries must be non-negative")
	}
	if c.PollEverySec <= 0 {
		return fmt.Errorf("poll step must be positive")
	}
	return nil
}
