package cache

import (
	"sync"
	"time"
)

// CacheValue 是一个接口
type CacheValue interface {
	Expiry() time.Time
}

// Cache[T CacheValue] 是一个泛型缓存
type Cache[T CacheValue] struct {
	// 并发安全的map
	data sync.Map
	// 清理过期数据的时间间隔
	cleanup time.Duration
}

func NewCache[T CacheValue](cleanupInterval time.Duration) *Cache[T] {
	cache := &Cache[T]{
		cleanup: cleanupInterval,
	}
	// 启动缓存清理 goroutine
	if cleanupInterval > 0 {
		go cache.startCleanup()
	}
	return cache
}

func (c *Cache[T]) Set(key any, value T) {
	c.data.Store(key, value)
}
func (c *Cache[T]) Get(key any) (value any, ok bool) {
	return c.data.Load(key)
}
func (c *Cache[T]) Delete(key any) {
	c.data.Delete(key)
}

// IsValueExpired 判断缓存值是否过期
//
// Return
//   - {bool} true 过期，false 未过期
func (c *Cache[T]) IsValueExpired(value T) bool {
	return time.Now().After(value.Expiry())
}

// IsExpired 通过键判断缓存值是否过期
//
// Return
//   - {bool} true 过期，false 未过期
func (c *Cache[T]) IsExpired(key any) bool {
	if value, ok := c.Get(key); ok {
		if vc, ok := value.(T); ok {
			return c.IsValueExpired(vc) // false 未过期 true 过期
		}
	}
	return false // 缓存中不存在该键(或值类型不匹配)
}

func (c *Cache[T]) startCleanup() {
	ticker := time.NewTicker(c.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		c.data.Range(func(key, value any) bool {
			if vc, ok := value.(T); ok {
				if c.IsValueExpired(vc) {
					c.Delete(key)
				}
			}
			return true
		})
	}
}
