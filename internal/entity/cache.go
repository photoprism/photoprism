package entity

import (
	"strings"
	"sync"
)

type CacheItem interface {
	CacheUID() string
	CacheName() string
}

type CacheItems []CacheItem

type Cached map[string]CacheItem

// Names maps names with unique ids.
func (c Cached) Names() map[string]string {
	names := make(map[string]string, len(c))

	for uid := range c {
		item := c[uid].(CacheItem)

		if name := item.CacheName(); name != "" {
			names[strings.ToLower(name)] = item.CacheUID()
		}
	}

	return names
}

// Cache represents a lightweight entity cache.
type Cache struct {
	count int
	items Cached
	names map[string]string
	mutex sync.RWMutex
}

// NewCache creates a new NewCache instance.
func NewCache(items Cached) *Cache {
	m := &Cache{
		items: items,
		count: len(items),
		names: items.Names(),
		mutex: sync.RWMutex{},
	}

	return m
}

// Count returns the number of cached items.
func (c *Cache) Count() int {
	return c.count
}

// Set adds or updates an item.
func (c *Cache) Set(item CacheItem) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	uid := item.CacheUID()
	name := strings.ToLower(item.CacheName())

	// Update name index.
	if o, found := c.items[uid]; !found {
		if name != "" {
			c.names[name] = uid
		}
	} else if n := strings.ToLower(o.(CacheItem).CacheName()); n != name {
		delete(c.names, n)
	}

	// Set item.
	c.items[uid] = item

	// Update count.
	c.count = len(c.items)
}

// UID finds and returns a cached item by unique id.
func (c *Cache) UID(uid string) (item CacheItem, found bool) {
	if c.count == 0 || uid == "" {
		return nil, false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if result, ok := c.items[uid]; ok && result != nil {
		return result, true
	} else {
		return nil, false
	}
}

// Name finds and returns a cached item by common name.
func (c *Cache) Name(name string) (item CacheItem, found bool) {
	if c.count == 0 || name == "" {
		return nil, false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	name = strings.ToLower(name)

	if uid, hasUid := c.names[name]; !hasUid {
		return nil, false
	} else if result, ok := c.items[uid]; ok && result != nil {
		return result, true
	} else {
		return nil, false
	}
}

// Remove a file from the lookup table.
func (c *Cache) Remove(uid string) {
	if c.count == 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Delete from name index.
	if item, found := c.items[uid]; !found || item == nil {
		return
	} else if name := strings.ToLower(item.(CacheItem).CacheName()); name != "" {
		delete(c.names, name)
	}

	// Delete from items.
	delete(c.items, uid)

	// Update count.
	c.count = len(c.items)
}

// Exists tests of a file exists.
func (c *Cache) Exists(uid string) bool {
	if c.count == 0 {
		return false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if _, ok := c.items[uid]; ok {
		return true
	} else {
		return false
	}
}

// NameExists checks if the item name is known.
func (c *Cache) NameExists(name string) bool {
	_, found := c.Name(name)
	return found
}
