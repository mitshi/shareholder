package main

import (
	"math/rand"
	"sync"
	"time"
)

const randomBytes = "1234567890"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randomBytes[rand.Intn(len(randomBytes))]
	}
	return string(b)
}

type OtpItem struct {
	Otp        string
	Expiration int64
}

func (item OtpItem) isExpired() bool {
	return time.Now().UnixNano() > item.Expiration
}

type OtpCache struct {
	items map[string]OtpItem
	mu    sync.RWMutex
}

func (c *OtpCache) CreateOtp(key string) string {
	expiration := time.Now().Add(5 * time.Minute).UnixNano()
	c.mu.Lock()
	if item, found := c.items[key]; found {
		return item.Otp
	}
	c.items[key] = OtpItem{
		Otp:        RandStringBytes(4),
		Expiration: expiration,
	}
	c.mu.Unlock()
	return c.items[key].Otp
}

func (c *OtpCache) GetOtp(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found || item.isExpired() {
		return "", false
	}
	return item.Otp, true
}

func (c *OtpCache) DeleteExpired() {
	for k, item := range c.items {
		if item.isExpired() {
			delete(c.items, k)
		}
	}
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *OtpCache) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
			break
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func NewOtpCache() *OtpCache {
	c := &OtpCache{
		items: make(map[string]OtpItem),
	}
	j := &janitor{
		Interval: time.Minute,
		stop:     make(chan bool),
	}
	go j.Run(c)
	return c
}
