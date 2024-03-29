package lru

import "container/list"

type Cache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

func New(maxBytes int64, OnEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if val, ok := c.cache[key]; ok {
		c.ll.MoveToFront(val)
		kv := val.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldList() {
	val := c.ll.Back()
	if val != nil {
		c.ll.Remove(val)

		kv := val.Value.(*entry)
		delete(c.cache, kv.key)

		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, val Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(val.Len()) - int64(kv.value.Len())
		kv.value = val
	} else {
		ele := c.ll.PushFront(&entry{key, val})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(val.Len())

	}
	// remove out if space is not enough
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldList()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
