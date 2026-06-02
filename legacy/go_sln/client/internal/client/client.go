package client

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sln/client/internal/config"
	"sln/client/internal/frame"
	"sln/client/internal/util"
	"sync"
	"time"
)

// Client отвечает за подключение к эмулятору и периодический опрос времени
type Client struct {
	cfg    *config.Config
	logger *log.Logger

	mu       sync.Mutex
	conn     net.Conn
	stopCh   chan struct{}
	running  bool
	lastSec  int
	wg       sync.WaitGroup
	dialLock sync.Mutex
}

// NewClient создаёт новый клиент с конфигом и логгером
func NewClient(cfg *config.Config, logger *log.Logger) *Client {
	return &Client{
		cfg:     cfg,
		logger:  logger,
		stopCh:  make(chan struct{}),
		lastSec: -1,
	}
}

// Start пытается подключиться и запускает цикл опроса (в фоне).
func (c *Client) Start() error {
	_ = c.reconnect()

	c.running = true
	c.wg.Add(1)
	go c.pollLoop()
	return nil
}

// Stop корректно останавливает клиент: закрывает соединение и ждёт горутин
func (c *Client) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	close(c.stopCh)
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.mu.Unlock()
	c.wg.Wait()
}

// connect устанавливает TCP-соединение к серверу
func (c *Client) connect() error {
	c.dialLock.Lock()
	defer c.dialLock.Unlock()

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = conn
	c.mu.Unlock()
	c.logger.Printf("connected to %s", addr)
	return nil
}

// pollLoop - основной цикл, тикер каждую секунду; запускает опрос на секундах кратных 5
func (c *Client) pollLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(time.Duration(c.cfg.PollEverySec) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			c.logger.Printf("poll loop stopping")
			return
		case now := <-ticker.C:
			sec := now.Second()
			if sec%5 == 0 && sec != c.lastSec {
				c.lastSec = sec
				c.wg.Add(1)
				go func() {
					defer c.wg.Done()
					c.performPoll()
				}()
			}
			if sec%5 != 0 {
				c.lastSec = -1
			}
		}
	}
}

// performPoll формирует запрос, отправляет и обрабатывает ответ с retry/timeout
func (c *Client) performPoll() {
	if err := c.ensureConn(); err != nil {
		c.logger.Printf("cannot connect: %v", err)
		return
	}

	control := byte(0x00)
	addr := byte(c.cfg.AdapterAddr & 0xFF)
	data := []byte{0x01} // команда чтения времени

	req := frame.BuildSkeleton(control, addr, data)
	req = frame.AppendChecksum(req, c.cfg.CRCMode)

	c.logger.Printf("TX request: %s", util.HexDump(req))

	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			c.logger.Printf("retry attempt %d", attempt)
		}
		if err := c.write(req); err != nil {
			lastErr = err
			c.logger.Printf("write error: %v", err)
			_ = c.reconnect()
			time.Sleep(200 * time.Millisecond)
			continue
		}
		resp, err := c.readFrameWithTimeout(time.Duration(c.cfg.TimeoutMs) * time.Millisecond)
		if err != nil {
			lastErr = err
			c.logger.Printf("read error: %v", err)
			_ = c.reconnect()
			time.Sleep(200 * time.Millisecond)
			continue
		}
		c.logger.Printf("RX response: %s", util.HexDump(resp))

		// Проверка контрольной суммы/структуры фрейма
		if err := frame.VerifyFrame(resp); err != nil {
			lastErr = err
			c.logger.Printf("frame verification failed: %v", err)
			continue
		}
		payload := frame.PayloadData(resp)
		if len(payload) == 0 {
			c.logger.Printf("empty payload")
			return
		}
		if payload[0] != 0x01 {
			c.logger.Printf("unexpected cmd in payload: 0x%02X", payload[0])
			return
		}
		timeStr := string(payload[1:])
		ts, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			// Если парсинг не удаётся - логируем raw строку
			c.logger.Printf("time parse failed, raw='%s'", timeStr)
			c.logger.Printf("device time (raw): %s", timeStr)
		} else {
			c.logger.Printf("device time: %s", ts.Format(time.RFC3339))
		}
		return
	}
	c.logger.Printf("all retries failed: last error: %v", lastErr)
}

// write отправляет байты в текущее соединение (защищено мьютексом).
func (c *Client) write(b []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("no connection")
	}
	_, err := c.conn.Write(b)
	return err
}

// readFrameWithTimeout читает данные из соединения до образования полного фрейма или таймаута
func (c *Client) readFrameWithTimeout(timeout time.Duration) ([]byte, error) {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("no connection")
	}
	conn := c.conn
	c.mu.Unlock()

	var buf bytes.Buffer
	tmp := make([]byte, 1024)
	deadline := time.Now().Add(timeout)

	for {
		_ = conn.SetReadDeadline(deadline)
		n, err := conn.Read(tmp)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if frameBytes, ok := frame.ExtractFrame(&buf); ok {
			return frameBytes, nil
		}
	}
}

// ensureConn убеждается, что есть открытое соединение, иначе пытается reconnect
func (c *Client) ensureConn() error {
	c.mu.Lock()
	if c.conn != nil {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()
	return c.reconnect()
}

// reconnect переподключается к серверу (с блокировкой, чтобы не было parallel dial).
func (c *Client) reconnect() error {
	c.dialLog("reconnecting...")
	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		c.dialLog("reconnect failed: %v", err)
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.dialLog("reconnected")
	return nil
}

// dialLog - вспомогательный лог для событий подключения
func (c *Client) dialLog(format string, args ...interface{}) {
	c.logger.Printf("[dial] "+format, args...)
}
