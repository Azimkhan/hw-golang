package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type keyValue struct {
	key   Key
	value interface{}
}
type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mutex    sync.RWMutex
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	kv := &keyValue{key, value}
	if item, ok := c.items[key]; ok {
		item.Value = kv
		c.queue.MoveToFront(item)
		return true
	}
	item := c.queue.PushFront(kv)
	c.items[key] = item
	if c.queue.Len() > c.capacity {
		back := c.queue.Back()
		c.queue.Remove(back)
		pair := back.Value.(*keyValue)
		delete(c.items, pair.key)
	}

	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if item, ok := c.items[key]; ok {
		return item.Value.(*keyValue).value, true
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}
