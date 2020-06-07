package lru

import (
	"container/list"
	"errors"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// LRU implements a non-thread safe fixed size LRU cache
type LRU struct {
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
	tags      map[string]map[*list.Element]bool
	onEvict   EvictCallback
}

// entry is used to hold a value in the evictList
type entry struct {
	key   interface{}
	value interface{}
	tags  []string
}

// NewLRU constructs an LRU of the given size
func NewLRU(size int, onEvict EvictCallback) (*LRU, error) {
	if size <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	c := &LRU{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		tags:      make(map[string]map[*list.Element]bool),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LRU) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry).value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
	c.tags = make(map[string]map[*list.Element]bool)
}

// Add adds a value to the cache and registers the tags by which can be invalidated.
// Returns true if an eviction occurred.
func (c *LRU) Add(key, value interface{}, tags ...string) (evicted bool) {
	// Check for existing item
	if el, ok := c.items[key]; ok {
		c.evictList.MoveToFront(el)
		c.untag(el, tags)
		el.Value.(*entry).value = value
		el.Value.(*entry).tags = tags
		c.tag(el, tags)
		return false
	}

	// Add new item
	ent := &entry{key, value, tags}
	el := c.evictList.PushFront(ent)
	c.tag(el, tags)
	c.items[key] = el

	evict := c.evictList.Len() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(key interface{}) (value interface{}, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(key interface{}) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU) Peek(key interface{}) (value interface{}, ok bool) {
	var el *list.Element
	if el, ok = c.items[key]; ok {
		return el.Value.(*entry).value, true
	}
	return nil, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) Remove(key interface{}) (present bool) {
	if el, ok := c.items[key]; ok {
		c.removeElement(el)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	el := c.evictList.Back()
	if el != nil {
		c.removeElement(el)
		ent := el.Value.(*entry)
		return ent.key, ent.value, true
	}
	return nil, nil, false
}

// GetOldest returns the oldest entry
func (c *LRU) GetOldest() (key interface{}, value interface{}, ok bool) {
	el := c.evictList.Back()
	if el != nil {
		ent := el.Value.(*entry)
		return ent.key, ent.value, true
	}
	return nil, nil, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for el := c.evictList.Back(); el != nil; el = el.Prev() {
		keys[i] = el.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	return c.evictList.Len()
}

// Resize changes the cache size.
func (c *LRU) Resize(size int) (evicted int) {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

// removeOldest removes the oldest item from the cache.
func (c *LRU) removeOldest() {
	el := c.evictList.Back()
	if el != nil {
		c.removeElement(el)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *LRU) removeElement(el *list.Element) {
	c.evictList.Remove(el)
	ent := el.Value.(*entry)
	c.untag(el, ent.tags)
	delete(c.items, ent.key)
	if c.onEvict != nil {
		c.onEvict(ent.key, ent.value)
	}
}

// Invalidate invalidates a tag, purging all associated keys from the cache.
func (c *LRU) Invalidate(tags []string) (removed int) {
	for _, tag := range tags {
		if els, ok := c.tags[tag]; ok {
			for el, _ := range els {
				c.removeElement(el)
				removed++
			}
		}
	}
	return removed
}

// FindByTags returns all matching keys for a set of tags.
func (c *LRU) FindByTags(tags []string) (keys []interface{}) {
	keys = []interface{}{}
	for _, tag := range tags {
		if els, ok := c.tags[tag]; ok {
			for el, _ := range els {
				keys = append(keys, el.Value.(*entry).key)
			}
		}
	}
	return
}

// tag adds a key to the invalidation list
func (c *LRU) tag(el *list.Element, tags []string) {
	for _, tag := range tags {
		if _, ok := c.tags[tag]; !ok {
			c.tags[tag] = make(map[*list.Element]bool)
		}
		c.tags[tag][el] = true
	}
}

// untag removes a key from the invalidation list
func (c *LRU) untag(el *list.Element, tags []string) {
	for _, tag := range tags {
		// Unnecessary safety measure removed for performance (~5%)
		// If a tagged element exists, its tags can be expected to exist
		// if _, ok := c.tags[tag]; !ok {
			// continue
		// }
		delete(c.tags[tag], el)
		if len(c.tags[tag]) == 0 {
			delete(c.tags, tag)
		}
	}
}
