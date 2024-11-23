package cache

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

var (
	ErrNoKey = errors.New("no item with the specified key")
)

type Cache struct {
	items    map[string]*list.Element
	lru      *list.List
	stopChan chan any
	ttl      int64
	capacity int
	mu       sync.RWMutex
}

type cacheItem struct {
	key     string
	value   int
	expires int64 // unixtime.
}

func NewCache(ttl int64, capacity int) *Cache {
	cache := &Cache{
		items:    make(map[string]*list.Element),
		lru:      list.New(),
		ttl:      ttl,
		mu:       sync.RWMutex{},
		capacity: capacity,
		stopChan: make(chan any),
	}

	return cache
}

// Peek retrieves a value without updating its ttl.
func (cache *Cache) Peek(key string) (int, error) {
	var item cacheItem

	cache.mu.RLock()

	listItem, keyOk := cache.items[key]
	if keyOk {
		item = *listItem.Value.(*cacheItem)
	}

	cache.mu.RUnlock()

	if !keyOk {
		return item.value, ErrNoKey
	}

	if item.expires <= time.Now().Unix() {
		cache.mu.Lock()
		delete(cache.items, key)
		cache.lru.Remove(listItem)
		cache.mu.Unlock()

		return item.value, ErrNoKey
	}

	return item.value, nil
}

// Get retrieves a value and updates its ttl.
func (cache *Cache) Get(key string) (int, error) {
	var item cacheItem

	cache.mu.Lock()

	listItem, keyOk := cache.items[key]
	if keyOk {
		item = *listItem.Value.(*cacheItem)
		cache.lru.MoveToBack(listItem)
	}

	cache.mu.Unlock()

	if !keyOk {
		return item.value, ErrNoKey
	}

	if item.expires <= time.Now().Unix() {
		cache.mu.Lock()
		delete(cache.items, key)
		cache.lru.Remove(listItem)
		cache.mu.Unlock()

		return item.value, ErrNoKey
	}

	return item.value, nil
}

// Set sets a new value.
func (cache *Cache) Set(key string, value int) {
	//nolint:exhaustruct // the expiration will be filled later.
	newItem := cacheItem{
		value: value,
		key:   key,
	}

	cache.mu.Lock()

	cache.set(newItem)

	cache.mu.Unlock()
}

func (cache *Cache) set(newItem cacheItem) {
	for cache.lru.Len() > cache.capacity {
		front := cache.lru.Front()
		delete(cache.items, front.Value.(*cacheItem).key)
		cache.lru.Remove(front)
	}

	newItem.expires = time.Now().Unix() + cache.ttl

	if cache.items[newItem.key] != nil {
		cache.items[newItem.key].Value.(*cacheItem).value = newItem.value
		cache.items[newItem.key].Value.(*cacheItem).expires = newItem.expires
		cache.lru.MoveToBack(cache.items[newItem.key])
	} else {
		cache.items[newItem.key] = cache.lru.PushBack(&newItem)
	}
}

// GetAndIncrease retrieves a value and increases it in the cache with ttl update.
func (cache *Cache) GetAndIncrease(key string) int {
	newItem := cacheItem{ //nolint:exhaustruct // the 'expires' is not used.
		value: 1,
		key:   key,
	}

	cache.mu.Lock()

	listItem, keyOk := cache.items[key]

	// set value (1) if there was no such key, or if the entry has expired.
	if keyOk && listItem.Value.(*cacheItem).expires > time.Now().Unix() {
		newItem = *listItem.Value.(*cacheItem)
		listItem.Value.(*cacheItem).value++
		listItem.Value.(*cacheItem).expires = time.Now().Unix() + cache.ttl
		cache.lru.MoveToBack(listItem)
	} else {
		cache.set(newItem)
	}

	cache.mu.Unlock()

	return newItem.value
}

func (cache *Cache) StopAutoEvict() {
	cache.stopChan <- struct{}{}
}

func (cache *Cache) AutoEvict(period time.Duration) {
	if period <= 0 {
		return
	}

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cache.doAutoEvict()
		case <-cache.stopChan:
			return
		}
	}
}

func (cache *Cache) doAutoEvict() {
	cache.mu.Lock()

	listItem := cache.lru.Front()

	for listItem != nil {
		if listItem.Value.(*cacheItem).expires > time.Now().Unix() {
			break
		}

		nextItem := listItem.Next()

		delete(cache.items, listItem.Value.(*cacheItem).key)
		cache.lru.Remove(listItem)

		listItem = nextItem
	}

	cache.mu.Unlock()
}
