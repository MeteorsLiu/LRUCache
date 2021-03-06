package cache

import (
	"container/list"
	"sync"
	"time"
)

/*
Copyright 2013 Google Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package lru implements an LRU cache.

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key Key, value interface{})

	ll    *list.List
	cache map[interface{}]*list.Element
	//mutex does't require init
	mu sync.Mutex
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key    Key
	value  interface{}
	expire int64
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache) Set(key Key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	//the map type is not concurrency safe.
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{
		key:    key,
		value:  value,
		expire: 0,
	})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries+1 {
		c.RemoveOldest()
	}
}

func (c *Cache) SetWithExpire(key Key, value interface{}, expiretime time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	//the map type is not concurrency safe.
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{
		key:    key,
		value:  value,
		expire: time.Now().Add(expiretime).Unix(),
	})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries+1 {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	//Visit the member of struct is safe.
	//Don't worry about it.
	if c.cache == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// GetAndRemoveExpire loos up a key's value ,returns if it exists and call
// a defer func to check it whether it's expired or not.
// If it was expired,remove it
func (c *Cache) GetAndRemoveExpire(key Key) (value interface{}, ok bool) {
	//Visit the member of struct is safe.
	//Don't worry about it.
	if c.cache == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, hit := c.cache[key]; hit {
		if ele.Value.(*entry).expire > 0 {
			defer func() {
				if time.Now().Unix() >= ele.Value.(*entry).expire {
					//No need to lock this.
					//Because defer Unlock() wil run afer this function
					c.removeElement(ele)
					return
				}
			}()
		}
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

func (c *Cache) Has(key Key) (hit bool) {
	if c.cache == nil {
		return
	}
	//It's safe to read the map only.
	_, hit = c.cache[key]
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	c.mu.Lock()
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
	c.mu.Unlock()
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}

//Reset all cache value and clear all key.
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.cache {
		c.removeElement(e)
	}
}

func (c *Cache) RemoveExpire() {
	for _, e := range c.cache {
		if e.Value.(*entry).expire > 0 {
			if time.Now().Unix() >= e.Value.(*entry).expire {
				c.mu.Lock()
				c.removeElement(e)
				c.mu.Unlock()
			}
		}
	}
}
