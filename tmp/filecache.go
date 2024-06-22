package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	filePath  string
	saveMutex sync.Mutex
}

// NewCache 创建一个新的缓存对象
func NewCache(filePath string) *Cache {
	return &Cache{
		filePath: filePath,
	}
}

// Set 向缓存中添加一个项
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.saveMutex.Lock()
	defer c.saveMutex.Unlock()

	items, err := c.loadFromFile()
	if err != nil {
		fmt.Println("Error loading cache from file:", err)
		return
	}

	expiration := time.Now().Add(duration).UnixNano()
	items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}

	c.saveToFile(items)
}

// Get 从缓存中获取一个项
func (c *Cache) Get(key string) (interface{}, bool) {
	c.saveMutex.Lock()
	defer c.saveMutex.Unlock()

	items, err := c.loadFromFile()
	if err != nil {
		fmt.Println("Error loading cache from file:", err)
		return nil, false
	}

	item, found := items[key]
	if !found {
		return nil, false
	}

	// 检查是否过期
	if time.Now().UnixNano() > item.Expiration {
		delete(items, key)
		c.saveToFile(items)
		return nil, false
	}
	return item.Value, true
}

// Delete 从缓存中删除一个项
func (c *Cache) Delete(key string) {
	c.saveMutex.Lock()
	defer c.saveMutex.Unlock()

	items, err := c.loadFromFile()
	if err != nil {
		fmt.Println("Error loading cache from file:", err)
		return
	}

	delete(items, key)
	c.saveToFile(items)
}

// saveToFile 将缓存保存到文件
func (c *Cache) saveToFile(items map[string]CacheItem) {
	file, err := os.Create(c.filePath)
	if err != nil {
		fmt.Println("Error creating cache file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
		fmt.Println("Error encoding cache to file:", err)
	}
}

// loadFromFile 从文件加载缓存
func (c *Cache) loadFromFile() (map[string]CacheItem, error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 如果文件不存在，返回一个空的 map
			return make(map[string]CacheItem), nil
		}
		return nil, err
	}
	defer file.Close()

	items := make(map[string]CacheItem)
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func main() {
	// 创建缓存对象，指定缓存文件路径
	cache := NewCache("cache.json")

	cache.Set("key1", "value1", 5*time.Second)
	cache.Set("key2", "value2", 10*time.Second)

	time.Sleep(3 * time.Second)

	value, found := cache.Get("key1")
	if found {
		fmt.Println("key1:", value)
	} else {
		fmt.Println("key1 not found")
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
