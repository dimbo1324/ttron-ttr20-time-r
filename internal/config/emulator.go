package config

import (
	"flag"
	"fmt"
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

func LoadEmulator() *EmulatorConfig {
	c := &EmulatorConfig{}
	flag.StringVar(&c.Host, "host", "127.0.0.1", "listen host")
	flag.IntVar(&c.Port, "port", 9000, "listen port")
	flag.StringVar(&c.Listen, "listen", "", "listen address, overrides -host/-port when set")
	flag.StringVar(&c.CRCMode, "crc", "sum", "crc mode: sum | crc16")
	mode := flag.String("mode", "", "checksum mode alias: sum | crc16")
	flag.IntVar(&c.DelayMs, "delay", 0, "fixed delay before responding (ms)")
	flag.DurationVar(&c.ResponseDelay, "response-delay", 0, "fixed delay before responding")
	flag.Float64Var(&c.BadCRCProb, "badcrc", 0.0, "probability [0..1] to send bad CRC in responses")
	badChecksum := flag.Bool("bad-checksum", false, "always corrupt response checksum")
	flag.Float64Var(&c.FragProb, "fragment", 0.0, "probability [0..1] to fragment responses")
	fragmentResponse := flag.Bool("fragment-response", false, "always fragment responses")
	flag.DurationVar(&c.FragmentDelay, "fragment-delay", 40*time.Millisecond, "delay between fragmented response writes")
	flag.BoolVar(&c.NoResponse, "no-response", false, "receive valid request but do not send a response")
	flag.BoolVar(&c.CloseAfterRequest, "close-after-request", false, "close connection after processing a request")
	flag.IntVar(&c.AdapterAddr, "adapter", 1, "adapter address byte (0..255)")
	flag.StringVar(&c.LogFile, "log", "", "path to log file; empty = stdout")
	flag.StringVar(&c.GRPCListen, "grpc-listen", ":9100", "gRPC control listen address; empty disables gRPC")
	flag.IntVar(&c.ReadTimeout, "readtimeout", 300, "connection read timeout in seconds")
	flag.DurationVar(&c.ReadTimeoutDuration, "read-timeout", 300*time.Second, "connection read timeout")
	flag.DurationVar(&c.WriteTimeoutDuration, "write-timeout", 5*time.Second, "connection write timeout")
	flag.IntVar(&c.RecentSize, "recent", 100, "recent frame/event buffer size")
	flag.Parse()
	if *mode != "" {
		c.CRCMode = *mode
	}
	if *badChecksum {
		c.BadCRCProb = 1
	}
	if *fragmentResponse {
		c.FragProb = 1
	}
	if c.ResponseDelay == 0 && c.DelayMs > 0 {
		c.ResponseDelay = time.Duration(c.DelayMs) * time.Millisecond
	}
	if c.ReadTimeoutDuration == 300*time.Second && c.ReadTimeout != 300 {
		c.ReadTimeoutDuration = time.Duration(c.ReadTimeout) * time.Second
	}
	return c
}

func (c EmulatorConfig) ListenAddress() string {
	if c.Listen != "" {
		return c.Listen
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
