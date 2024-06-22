package main

import (
	"fmt"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
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
