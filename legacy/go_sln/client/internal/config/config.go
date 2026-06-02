package config

import "flag"

// Config хранит параметры запуска клиента
type Config struct {
	Host         string
	Port         int
	CRCMode      string
	AdapterAddr  int
	TimeoutMs    int
	Retries      int
	LogFile      string
	PollEverySec int
}

// Load парсит флаги командной строки и возвращает конфиг
func Load() *Config {
	c := &Config{}
	flag.StringVar(&c.Host, "host", "127.0.0.1", "server host")
	flag.IntVar(&c.Port, "port", 9000, "server port")
	flag.StringVar(&c.CRCMode, "crc", "sum", "crc mode: sum | crc16")
	flag.IntVar(&c.AdapterAddr, "adapter", 1, "adapter address (0..255)")
	flag.IntVar(&c.TimeoutMs, "timeout", 1000, "timeout for response in milliseconds")
	flag.IntVar(&c.Retries, "retries", 2, "number of retries on timeout/error")
	flag.StringVar(&c.LogFile, "log", "", "log file (empty = stdout)")
	flag.IntVar(&c.PollEverySec, "pollstep", 1, "polling tick step in seconds (default 1)")
	flag.Parse()
	return c
}
