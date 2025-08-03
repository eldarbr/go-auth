package cache

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

var (
	ErrNoKey = errors.New("no item with the specified key")
	ErrCast  = errors.New("internal error - list item cast failed")
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
		itemPtr, castOk := listItem.Value.(*cacheItem)
		if !castOk {
			cache.mu.RUnlock()

			return item.value, ErrCast
		}

		item = *itemPtr
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
		itemPtr, castOk := listItem.Value.(*cacheItem)
		if !castOk {
			cache.mu.Unlock()

			return item.value, ErrCast
		}

		item = *itemPtr

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

// concurrent-UNSAFE.
func (cache *Cache) set(newItem cacheItem) {
	var (
		frontPtr       *cacheItem
		frontPtrCastOk bool
	)

	for cache.lru.Len() > cache.capacity {
		frontList := cache.lru.Front()
		if frontList == nil {
			return
		} else if frontPtr, frontPtrCastOk = frontList.Value.(*cacheItem); !frontPtrCastOk {
			return
		}

		delete(cache.items, frontPtr.key)
		cache.lru.Remove(frontList)
	}

	newItem.expires = time.Now().Unix() + cache.ttl

	if cache.items[newItem.key] != nil {
		itemPtr, itemPtrCastOk := cache.items[newItem.key].Value.(*cacheItem)
		if !itemPtrCastOk {
			return
		}

		itemPtr.value = newItem.value
		itemPtr.expires = newItem.expires
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

	var currentItemPtr *cacheItem

	cache.mu.Lock()

	listItem, keyOk := cache.items[key]
	if keyOk {
		itemPtr, itemPtrCastOk := listItem.Value.(*cacheItem)
		if !itemPtrCastOk {
			cache.mu.Unlock()

			return newItem.value
		}

		currentItemPtr = itemPtr
	}

	// set value (1) if there was no such key, or if the entry has expired.
	if keyOk && currentItemPtr != nil && currentItemPtr.expires > time.Now().Unix() {
		newItem = *currentItemPtr
		currentItemPtr.value++
		currentItemPtr.expires = time.Now().Unix() + cache.ttl
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
			cache.DoAutoEvict()
		case <-cache.stopChan:
			return
		}
	}
}

func (cache *Cache) DoAutoEvict() {
	cache.mu.Lock()

	listItem := cache.lru.Front()

	for listItem != nil {
		item, itemCastOk := listItem.Value.(*cacheItem)
		if !itemCastOk {
			cache.mu.Unlock()

			return
		}

		if item.expires > time.Now().Unix() {
			break
		}

		nextItem := listItem.Next()

		delete(cache.items, item.key)
		cache.lru.Remove(listItem)

		listItem = nextItem
	}

	cache.mu.Unlock()
}
