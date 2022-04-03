package entity

import "github.com/photoprism/photoprism/pkg/rnd"

var subjectCache *Cache

// CacheUID returns the subject's UID.
func (m Subject) CacheUID() string {
	return m.SubjUID
}

// CacheName returns the subject's name.
func (m Subject) CacheName() string {
	return m.SubjName
}

// Cached finds a cached subject.
func (m Subject) Cached(key string) *Subject {
	if key == "" {
		return nil
	}

	result := Subject{}

	// Query cache.
	if cached, found := m.Cache().UID(key); found && rnd.IsPPID(key, 'j') {
		result = *cached.(*Subject)
	} else if cached, found := m.Cache().Name(key); found {
		result = *cached.(*Subject)
	} else {
		return nil
	}

	return &result
}

// UpdateCache updates cached values.
func (m Subject) UpdateCache() {
	Subject{}.Cache().Set(&m)
}

// Cache returns the cache instance.
func (m Subject) Cache() *Cache {
	if subjectCache != nil {
		return subjectCache
	}

	var items []Subject

	if err := UnscopedDb().Find(&items).Error; err != nil {
		log.Tracef("cache: %s", err)
		subjectCache = NewCache(make(Cached))
	} else {
		cached := make(Cached, len(items))

		for _, i := range items {
			cached[i.CacheUID()] = i
		}

		subjectCache = NewCache(cached)
	}

	return subjectCache
}
