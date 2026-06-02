package client

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
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

	req := frame.BuildSkeleton(0x00, byte(c.cfg.AdapterAddr&0xFF), []byte{0x01})
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

		if err := frame.Verify(resp); err != nil {
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
			c.logger.Printf("time parse failed, raw='%s'", timeStr)
			c.logger.Printf("device time (raw): %s", timeStr)
			return
		}
		c.logger.Printf("device time: %s", ts.Format(time.RFC3339))
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
		if f, ok := frame.ExtractFrame(&buf); ok {
			return f, nil
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
