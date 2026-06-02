package config

import "flag"

// содержит параметры запуска сервера
type Config struct {
	Host        string
	Port        int
	CRCMode     string
	DelayMs     int
	BadCRCProb  float64
	FragProb    float64
	AdapterAddr int
	LogFile     string
	ReadTimeout int // секунды для таймаута чтения соединения
}

// парсит флаги командной строки и возвращает конфигурацию
func Load() *Config {
	confRes := &Config{}

	flag.StringVar(&confRes.Host, "host", "127.0.0.1", "listen host")
	flag.IntVar(&confRes.Port, "port", 9000, "listen port")
	flag.StringVar(&confRes.CRCMode, "crc", "sum", "crc mode: sum | crc16")
	flag.IntVar(&confRes.DelayMs, "delay", 0, "fixed delay before responding (ms)")
	flag.Float64Var(&confRes.BadCRCProb, "badcrc", 0.0, "probability [0..1] to send bad CRC in responses")
	flag.Float64Var(&confRes.FragProb, "fragment", 0.0, "probability [0..1] to fragment responses")
	flag.IntVar(&confRes.AdapterAddr, "adapter", 1, "adapter address byte (0..255)")
	flag.StringVar(&confRes.LogFile, "log", "", "path to log file; empty = stdout")
	flag.IntVar(&confRes.ReadTimeout, "readtimeout", 300, "connection read timeout in seconds")

	flag.Parse()
	return confRes
}
