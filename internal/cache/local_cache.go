package cache

// cache value interface with generic type

// create a local cache wirth map  and allow generic key and values
func NewLocalCache[K comparable, V any]() *LocalCache[K, V] {
	return &LocalCache[K, V]{
		cache: make(map[K]V),
	}
}

type LocalCache[K comparable, V any] struct {
	cache map[K]V
}

func (c *LocalCache[K, V]) Set(key K, value V) {
	c.cache[key] = value
}

func (c *LocalCache[K, V]) Get(key K) (V, bool) {
	value, ok := c.cache[key]
	return value, ok
}
