package config

import "flag"

type EmulatorConfig struct {
	Host        string
	Port        int
	CRCMode     string
	DelayMs     int
	BadCRCProb  float64
	FragProb    float64
	AdapterAddr int
	LogFile     string
	ReadTimeout int
}

func LoadEmulator() *EmulatorConfig {
	c := &EmulatorConfig{}
	flag.StringVar(&c.Host, "host", "127.0.0.1", "listen host")
	flag.IntVar(&c.Port, "port", 9000, "listen port")
	flag.StringVar(&c.CRCMode, "crc", "sum", "crc mode: sum | crc16")
	flag.IntVar(&c.DelayMs, "delay", 0, "fixed delay before responding (ms)")
	flag.Float64Var(&c.BadCRCProb, "badcrc", 0.0, "probability [0..1] to send bad CRC in responses")
	flag.Float64Var(&c.FragProb, "fragment", 0.0, "probability [0..1] to fragment responses")
	flag.IntVar(&c.AdapterAddr, "adapter", 1, "adapter address byte (0..255)")
	flag.StringVar(&c.LogFile, "log", "", "path to log file; empty = stdout")
	flag.IntVar(&c.ReadTimeout, "readtimeout", 300, "connection read timeout in seconds")
	flag.Parse()
	return c
}
