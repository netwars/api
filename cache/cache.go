package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrNotFound is returned by Get if there is no data under given key.
	ErrNotFound = errors.New("store: not found")
)

// CacheOpts ...
type CacheOpts struct {
	Expiration time.Duration
	Interval   time.Duration
}

// Cache ...
type Cache struct {
	sync.RWMutex
	err          chan error
	notification chan int
	expiration   time.Duration
	interval     time.Duration
	rows         map[int]interface{}
	retrieved    map[int]bool
	tickers      map[int]*time.Ticker
	timers       map[int]*time.Timer
	cancel       map[int]chan struct{}
}

// NewCache ...
func NewCache(options CacheOpts) *Cache {
	return &Cache{
		rows:         make(map[int]interface{}),
		tickers:      make(map[int]*time.Ticker),
		timers:       make(map[int]*time.Timer),
		cancel:       make(map[int]chan struct{}),
		err:          make(chan error, 1),
		notification: make(chan int),
		expiration:   options.Expiration,
		interval:     options.Interval,
	}
}

// Notify ...
func (c *Cache) Notify() <-chan int {
	return c.notification
}

// Set ...
func (c *Cache) Set(key int, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if _, exists := c.rows[key]; exists {
		if cancel, exists := c.cancel[key]; exists {
			close(cancel)
		}
		c.delete(key)
	}

	c.rows[key] = value
	c.tickers[key] = time.NewTicker(c.interval)
	c.timers[key] = time.NewTimer(c.expiration)
	c.cancel[key] = make(chan struct{})

	go c.schedule(key)
}

// Get ...
func (c *Cache) Get(key int) interface{} {
	if row, exists := c.rows[key]; exists {
		if c.timers[key].Reset(c.expiration) {
			return row
		}
	}

	return nil
}

// SafeGet ...
func (c *Cache) SafeGet(key int) interface{} {
	c.RLock()
	defer c.RUnlock()

	return c.Get(key)
}

// Delete ...
func (c *Cache) Delete(id int) {
	c.Lock()
	defer c.Unlock()

	if cancel, exists := c.cancel[id]; exists {
		close(cancel)
	}

	c.delete(id)
}

// Len ...
func (c *Cache) Len() int {
	c.RLock()
	defer c.RUnlock()

	return len(c.rows)
}

// Terminate ...
func (c *Cache) Terminate() {
	c.Lock()
	defer c.Unlock()

	for id := range c.rows {
		c.delete(id)
	}

	close(c.notification)
	close(c.err)
}

func (c *Cache) delete(key int) {
	if timer, exists := c.timers[key]; exists {
		timer.Stop()
	}

	if ticker, exists := c.tickers[key]; exists {
		ticker.Stop()
	}

	delete(c.rows, key)
	delete(c.cancel, key)
	delete(c.tickers, key)
	delete(c.timers, key)
}

// Err ...
func (c *Cache) Err() <-chan error {
	return c.err
}

func (c *Cache) schedule(id int) {
	c.RLock()
	ticker, exists := c.tickers[id]
	timer, exists2 := c.timers[id]
	cancel, exists3 := c.cancel[id]
	if !exists || !exists2 || !exists3 {
		c.RUnlock()
		return
	}
	c.RUnlock()

	for {
		select {
		case <-ticker.C:
			c.notification <- id
		case <-timer.C:
			c.Lock()
			c.delete(id)
			c.Unlock()

			c.err <- fmt.Errorf("Cache expired for ID: %d", id)
			return
		case <-cancel:
			return
		}
	}
}
