package cache

type Key string

type listValue struct {
	key   Key
	value interface{}
}

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	item, ok := c.items[key]
	itemValue := listValue{key: key, value: value}
	if ok {
		item.Value = itemValue
		c.queue.MoveToFront(item)
	} else {
		item = c.queue.PushFront(itemValue)
		c.items[key] = item
		if c.queue.Len() > c.capacity {
			lastElement := c.queue.Back()
			delete(c.items, lastElement.Value.(listValue).key)
			c.queue.Remove(lastElement)
		}
	}

	return ok
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := c.items[key]; ok {
		c.queue.MoveToFront(item)
		return item.Value.(listValue).value, ok
	}

	return nil, false
}

func (c *lruCache) Clear() {
	for k, v := range c.items {
		c.queue.Remove(v)
		delete(c.items, k)
	}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func NewCustomDeleter(deleter func(value interface{})) {
	customDeleter = deleter
}

func ResetDeleter() {
	customDeleter = nil
}
