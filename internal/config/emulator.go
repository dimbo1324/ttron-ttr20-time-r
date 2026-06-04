package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type EmulatorConfig struct {
	Host                 string
	Port                 int
	Listen               string
	CRCMode              string
	DelayMs              int
	ResponseDelay        time.Duration
	BadCRCProb           float64
	FragProb             float64
	FragmentDelay        time.Duration
	NoResponse           bool
	CloseAfterRequest    bool
	AdapterAddr          int
	LogFile              string
	GRPCListen           string
	ReadTimeout          int
	ReadTimeoutDuration  time.Duration
	WriteTimeoutDuration time.Duration
	RecentSize           int
}

func DefaultEmulator() EmulatorConfig {
	return EmulatorConfig{
		Host:                 "127.0.0.1",
		Port:                 9000,
		CRCMode:              "sum",
		FragmentDelay:        40 * time.Millisecond,
		AdapterAddr:          1,
		LogFile:              "runtime/logs/ft12-emulator.log",
		GRPCListen:           ":9100",
		ReadTimeout:          300,
		ReadTimeoutDuration:  300 * time.Second,
		WriteTimeoutDuration: 5 * time.Second,
		RecentSize:           100,
	}
}

func LoadEmulator(args []string) (*EmulatorConfig, error) {
	c := DefaultEmulator()
	fs := flag.NewFlagSet("ft12-emulator", flag.ContinueOnError)
	fs.StringVar(&c.Host, "host", c.Host, "listen host")
	fs.IntVar(&c.Port, "port", c.Port, "listen port")
	fs.StringVar(&c.Listen, "listen", c.Listen, "listen address, overrides -host/-port when set")
	fs.StringVar(&c.CRCMode, "crc", c.CRCMode, "crc mode: sum | crc16")
	mode := fs.String("mode", "", "checksum mode alias: sum | crc16")
	fs.IntVar(&c.DelayMs, "delay", c.DelayMs, "fixed delay before responding (ms)")
	fs.DurationVar(&c.ResponseDelay, "response-delay", c.ResponseDelay, "fixed delay before responding")
	fs.Float64Var(&c.BadCRCProb, "badcrc", c.BadCRCProb, "probability [0..1] to send bad CRC in responses")
	badChecksum := fs.Bool("bad-checksum", false, "always corrupt response checksum")
	fs.Float64Var(&c.FragProb, "fragment", c.FragProb, "probability [0..1] to fragment responses")
	fragmentResponse := fs.Bool("fragment-response", false, "always fragment responses")
	fs.DurationVar(&c.FragmentDelay, "fragment-delay", c.FragmentDelay, "delay between fragmented response writes")
	fs.BoolVar(&c.NoResponse, "no-response", c.NoResponse, "receive valid request but do not send a response")
	fs.BoolVar(&c.CloseAfterRequest, "close-after-request", c.CloseAfterRequest, "close connection after processing a request")
	fs.IntVar(&c.AdapterAddr, "adapter", c.AdapterAddr, "adapter address byte (0..255)")
	fs.StringVar(&c.LogFile, "log", c.LogFile, "path to log file; empty = stdout")
	fs.StringVar(&c.GRPCListen, "grpc-listen", c.GRPCListen, "gRPC control listen address; empty disables gRPC")
	fs.IntVar(&c.ReadTimeout, "readtimeout", c.ReadTimeout, "connection read timeout in seconds")
	fs.DurationVar(&c.ReadTimeoutDuration, "read-timeout", c.ReadTimeoutDuration, "connection read timeout")
	fs.DurationVar(&c.WriteTimeoutDuration, "write-timeout", c.WriteTimeoutDuration, "connection write timeout")
	fs.IntVar(&c.RecentSize, "recent", c.RecentSize, "recent frame/event buffer size")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *mode != "" {
		c.CRCMode = *mode
	}
	if *badChecksum {
		c.BadCRCProb = 1
	}
	if *fragmentResponse {
		c.FragProb = 1
	}
	c.Normalize()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func MustLoadEmulatorFromOS() *EmulatorConfig {
	cfg, err := LoadEmulator(os.Args[1:])
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *EmulatorConfig) Normalize() {
	if c.CRCMode == "" {
		c.CRCMode = "sum"
	}
	if c.ResponseDelay == 0 && c.DelayMs > 0 {
		c.ResponseDelay = time.Duration(c.DelayMs) * time.Millisecond
	}
	if c.ReadTimeoutDuration == 300*time.Second && c.ReadTimeout != 300 {
		c.ReadTimeoutDuration = time.Duration(c.ReadTimeout) * time.Second
	}
}

func (c EmulatorConfig) Validate() error {
	if err := validateTCPAddress(c.ListenAddress(), "listen address"); err != nil {
		return err
	}
	if err := validateAdapterAddr(c.AdapterAddr); err != nil {
		return err
	}
	if err := validateChecksumMode(c.CRCMode); err != nil {
		return err
	}
	if err := validateProbability(c.BadCRCProb, "bad CRC probability"); err != nil {
		return err
	}
	if err := validateProbability(c.FragProb, "fragment probability"); err != nil {
		return err
	}
	if err := validateRecentSize(c.RecentSize); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.ReadTimeoutDuration, "read timeout"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.WriteTimeoutDuration, "write timeout"); err != nil {
		return err
	}
	if c.GRPCListen != "" {
		if err := validateTCPAddress(c.GRPCListen, "gRPC listen address"); err != nil {
			return err
		}
	}
	if err := validateNonNegativeDuration(c.ResponseDelay, "response delay"); err != nil {
		return err
	}
	if err := validateNonNegativeDuration(c.FragmentDelay, "fragment delay"); err != nil {
		return err
	}
	return nil
}

func (c EmulatorConfig) ListenAddress() string {
	if c.Listen != "" {
		return c.Listen
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func validateTCPAddress(address, name string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%s must be host:port: %w", name, err)
	}
	if host != "" {
		if ip := net.ParseIP(host); ip == nil && !isDNSLikeHost(host) {
			return fmt.Errorf("%s host is invalid: %s", name, host)
		}
	}
	if port == "" {
		return fmt.Errorf("%s port must not be empty", name)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("%s port must be in range 1..65535", name)
	}
	return nil
}

func isDNSLikeHost(host string) bool {
	for _, r := range host {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return host != ""
}
