// Copyright (c) 2012-2019 Patrick Mylund Nielsen and the go-cache contributors
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2022 The FluxCD contributors. All rights reserved.
// This package provides an in-memory cache
// derived from the https://github.com/patrickmn/go-cache
// package
// It has been modified in order to keep a small set of functions
// and to add a maxItems parameter in order to limit the number of,
// and thus the size of, items in the cache.

package cache

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Cache is a thread-safe in-memory key/value store.
type Cache struct {
	*cache
}

// Item is an item stored in the cache.
type Item struct {
	// Object is the item's value.
	Object interface{}
	// Expiration is the item's expiration time.
	Expiration int64
}

type cache struct {
	// Items holds the elements in the cache.
	Items map[string]Item
	// MaxItems is the maximum number of items the cache can hold.
	MaxItems int
	mu       sync.RWMutex
	janitor  *janitor
}

// ItemCount returns the number of items in the cache.
// This may include items that have expired, but have not yet been cleaned up.
func (c *cache) ItemCount() int {
	c.mu.RLock()
	n := len(c.Items)
	c.mu.RUnlock()
	return n
}

func (c *cache) set(key string, value interface{}, expiration time.Duration) {
	var e int64
	if expiration > 0 {
		e = time.Now().Add(expiration).UnixNano()
	}

	c.Items[key] = Item{
		Object:     value,
		Expiration: e,
	}
}

// Set adds an item to the cache, replacing any existing item.
// If expiration is zero, the item never expires.
// If the cache is full, Set will return an error.
func (c *cache) Set(key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	_, found := c.Items[key]
	if found {
		c.set(key, value, expiration)
		c.mu.Unlock()
		return nil
	}

	if c.MaxItems > 0 && len(c.Items) < c.MaxItems {
		c.set(key, value, expiration)
		c.mu.Unlock()
		return nil
	}

	c.mu.Unlock()
	return fmt.Errorf("Cache is full")
}

// Add an item to the cache, existing items will not be overwritten.
// To overwrite existing items, use Set.
// If the cache is full, Add will return an error.
func (c *cache) Add(key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	_, found := c.Items[key]
	if found {
		c.mu.Unlock()
		return fmt.Errorf("Item %s already exists", key)
	}

	if c.MaxItems > 0 && len(c.Items) < c.MaxItems {
		c.set(key, value, expiration)
		c.mu.Unlock()
		return nil
	}

	c.mu.Unlock()
	return fmt.Errorf("Cache is full")
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.Items[key]
	if !found {
		c.mu.RUnlock()
		return nil, false
	}
	if item.Expiration > 0 {
		if item.Expiration < time.Now().UnixNano() {
			c.mu.RUnlock()
			return nil, false
		}
	}
	c.mu.RUnlock()
	return item.Object, true
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (c *cache) Delete(key string) {
	c.mu.Lock()
	delete(c.Items, key)
	c.mu.Unlock()
}

// Clear all items from the cache.
// This reallocate the inderlying array holding the items,
// so that the memory used by the items is reclaimed.
func (c *cache) Clear() {
	c.mu.Lock()
	c.Items = make(map[string]Item)
	c.mu.Unlock()
}

// HasExpired returns true if the item has expired.
func (c *cache) HasExpired(key string) bool {
	c.mu.RLock()
	item, ok := c.Items[key]
	if !ok {
		c.mu.RUnlock()
		return true
	}
	if item.Expiration > 0 {
		if item.Expiration < time.Now().UnixNano() {
			c.mu.RUnlock()
			return true
		}
	}
	c.mu.RUnlock()
	return false
}

// SetExpiration sets the expiration for the given key.
// Does nothing if the key is not in the cache.
func (c *cache) SetExpiration(key string, expiration time.Duration) {
	c.mu.Lock()
	item, ok := c.Items[key]
	if !ok {
		c.mu.Unlock()
		return
	}
	item.Expiration = time.Now().Add(expiration).UnixNano()
	c.mu.Unlock()
}

// GetExpiration returns the expiration for the given key.
// Returns zero if the key is not in the cache or the item
// has already expired.
func (c *cache) GetExpiration(key string) time.Duration {
	c.mu.RLock()
	item, ok := c.Items[key]
	if !ok {
		c.mu.RUnlock()
		return 0
	}
	if item.Expiration > 0 {
		if item.Expiration < time.Now().UnixNano() {
			c.mu.RUnlock()
			return 0
		}
	}
	c.mu.RUnlock()
	return time.Duration(item.Expiration - time.Now().UnixNano())
}

// DeleteExpired deletes all expired items from the cache.
func (c *cache) DeleteExpired() {
	c.mu.Lock()
	for k, v := range c.Items {
		if v.Expiration > 0 && v.Expiration < time.Now().UnixNano() {
			delete(c.Items, k)
		}
	}
	c.mu.Unlock()
}

type janitor struct {
	interval time.Duration
	stop     chan bool
}

func (j *janitor) run(c *cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}

// New creates a new cache with the given configuration.
func New(maxItems int, interval time.Duration) *Cache {
	c := &cache{
		Items:    make(map[string]Item),
		MaxItems: maxItems,
		janitor: &janitor{
			interval: interval,
			stop:     make(chan bool),
		},
	}

	C := &Cache{c}

	if interval > 0 {
		go c.janitor.run(c)
		runtime.SetFinalizer(C, stopJanitor)
	}

	return C
}
