package client

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

type Client struct {
	cfg    *config.ClientConfig
	logger *log.Logger

	mu       sync.Mutex
	conn     net.Conn
	stopCh   chan struct{}
	running  bool
	lastSec  int
	wg       sync.WaitGroup
	dialLock sync.Mutex
}

func New(cfg *config.ClientConfig, logger *log.Logger) *Client {
	return &Client{
		cfg:     cfg,
		logger:  logger,
		stopCh:  make(chan struct{}),
		lastSec: -1,
	}
}

func (c *Client) Start() error {
	_ = c.reconnect()
	c.running = true
	c.wg.Add(1)
	go c.pollLoop()
	return nil
}

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

func (c *Client) performPoll() {
	if err := c.ensureConn(); err != nil {
		c.logger.Printf("cannot connect: %v", err)
		return
	}

	mode, err := checksum.ParseMode(c.cfg.CRCMode)
	if err != nil {
		c.logger.Printf("invalid checksum mode: %v", err)
		return
	}
	wire := codec.New(mode, 0x00, byte(c.cfg.AdapterAddr&0xFF))
	req, err := wire.EncodeReadTimeRequest()
	if err != nil {
		c.logger.Printf("cannot encode read-time request: %v", err)
		return
	}
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
		resp, err := c.readFrameWithTimeout(mode, time.Duration(c.cfg.TimeoutMs)*time.Millisecond)
		if err != nil {
			lastErr = err
			c.logger.Printf("read error: %v", err)
			_ = c.reconnect()
			time.Sleep(200 * time.Millisecond)
			continue
		}
		c.logger.Printf("RX response: %s", util.HexDump(resp))

		_, parsed, err := wire.DecodeReadTimeResponse(resp)
		if err != nil {
			lastErr = err
			c.logger.Printf("read-time response parse failed: %v", err)
			continue
		}
		c.logger.Printf("device time: %s", parsed.Time.Format(time.RFC3339))
		return
	}
	c.logger.Printf("all retries failed: last error: %v", lastErr)
}

func (c *Client) write(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("no connection")
	}
	_, err := c.conn.Write(data)
	return err
}

func (c *Client) readFrameWithTimeout(mode checksum.Mode, timeout time.Duration) ([]byte, error) {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("no connection")
	}
	conn := c.conn
	c.mu.Unlock()

	parser := frame.NewStreamParser(mode)
	tmp := make([]byte, 1024)
	deadline := time.Now().Add(timeout)

	for {
		_ = conn.SetReadDeadline(deadline)
		n, err := conn.Read(tmp)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			continue
		}
		result := parser.Push(tmp[:n])
		for _, parseErr := range result.Errors {
			c.logger.Printf("protocol parse error: %v", parseErr)
		}
		if len(result.Frames) > 0 {
			return result.Frames[0].RawBytes(), nil
		}
	}
}

func (c *Client) ensureConn() error {
	c.mu.Lock()
	hasConn := c.conn != nil
	c.mu.Unlock()
	if hasConn {
		return nil
	}
	return c.reconnect()
}

func (c *Client) reconnect() error {
	c.dialLock.Lock()
	defer c.dialLock.Unlock()

	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	c.logger.Printf("[dial] reconnecting to %s...", addr)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		c.logger.Printf("[dial] reconnect failed: %v", err)
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.logger.Printf("[dial] reconnected")
	return nil
}
