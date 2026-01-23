package hw04lrucache

import "sync"

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type Key string

type lruCache struct {
	capacity int
	queue    List
	mu       sync.Mutex
	items    map[Key]*ListItem
}

type cacheItem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		// mu уже готов к использованию, после объявления присаиваются нулевые значения.
		items: make(map[Key]*ListItem, capacity),
	}
}

// Добавить значение в кэш по ключу.
func (cache *lruCache) Set(key Key, value interface{}) bool {
	var wasInCache bool
	var itemValue cacheItem

	cache.mu.Lock()         // залочили на запись.
	defer cache.mu.Unlock() // разблочим после.

	itemValue.key = key
	itemValue.value = value

	item, wasInCache := cache.items[key]

	if wasInCache {
		// обновить значение.
		item.Value = itemValue

		// переместить элемент в начало очереди.
		cache.queue.MoveToFront(item)
	} else {
		// новый элемент.

		// поместить в начало очереди.
		cache.queue.PushFront(itemValue)

		// добавить в словарь
		cache.items[key] = cache.queue.Front()

		if cache.queue.Len() > cache.capacity {
			// удалить из словаря по ключу.
			if v, ok := cache.queue.Back().Value.(cacheItem); ok {
				delete(cache.items, v.key)
			}

			// удалить последний элемент из очереди.
			cache.queue.Remove(cache.queue.Back())
		}
	}

	return wasInCache
}

// Получить значение из кэша по ключу.
func (cache *lruCache) Get(key Key) (interface{}, bool) {
	cache.mu.Lock()         // залочили на чтение.
	defer cache.mu.Unlock() // отпустили после.

	item, wasInCache := cache.items[key]

	if wasInCache {
		cache.queue.MoveToFront(item)
		if v, ok := cache.queue.Front().Value.(cacheItem); ok {
			return v.value, true
		}
	}
	return nil, false
}

// Очистить кэш.
func (cache *lruCache) Clear() {
	cache.mu.Lock()
	// новые пустые переменные.
	cache.items = make(map[Key]*ListItem, cache.capacity)
	cache.queue = NewList()
	cache.mu.Unlock()
}
