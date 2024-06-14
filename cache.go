package main

import (
	"fmt"
	"sync"
	"time"
)

// CacheItem 表示一个缓存项
type CacheItem struct {
	Value      interface{}
	Expiration int64
}

// Cache 表示缓存对象
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

// NewCache 创建一个新的缓存对象
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}
	// 启动一个 goroutine 定期清除过期项
	//go cache.cleanup()
	return cache
}

// Set 向缓存中添加一个项
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Get 从缓存中获取一个项
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	// 检查是否过期
	if time.Now().UnixNano() > item.Expiration {
		delete(c.items, key)
		return nil, false
	}
	return item.Value, true
}

// Delete 从缓存中删除一个项
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// cleanup 定期清除过期项
func (c *Cache) cleanup() {
	for {
		time.Sleep(2 * time.Hour) // 每2小时清理一次
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func main() {
	cache := NewCache()

	cache.Set("key1", "value1", 5*time.Second)
	cache.Set("key2", "value2", 10*time.Second)
	cache.Set("key3", map[string]interface{}{
		"txt":    "hhh",
		"status": 1,
	}, 10*time.Second)

	time.Sleep(3 * time.Second)

	value, found := cache.Get("key1")
	if found {
		fmt.Println("key1:", value)
	} else {
		fmt.Println("key1 not found")
	}
	value, found = cache.Get("key3")
	if found {
		fmt.Println("key3:", value, value.(map[string]interface{})["txt"])
	} else {
		fmt.Println("key3 not found")
	}

	time.Sleep(3 * time.Second)

	value, found = cache.Get("key1")
	if found {
		fmt.Println("key1:", value)
	} else {
		fmt.Println("key1 not found")
	}

	value, found = cache.Get("key2")
	if found {
		fmt.Println("key2:", value)
	} else {
		fmt.Println("key2 not found")
	}
}
