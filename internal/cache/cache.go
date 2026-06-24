package cache

import (
	"fmt"
	"sync"
	"time"
)

var store sync.Map

// Entry represents a cached value with an expiration time.
type Entry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// TTL is the time-to-live for cache entries.
const TTL = 30 * time.Second

// Get retrieves a value from the cache if it exists and has not expired.
func Get(key string) (interface{}, bool) {
	if val, ok := store.Load(key); ok {
		entry, ok := val.(Entry)
		if !ok {
			return nil, false
		}
		if time.Now().Before(entry.ExpiresAt) {
			return entry.Data, true
		}
		// Atomically delete only if the entry hasn't been refreshed by another goroutine
		store.CompareAndDelete(key, val)
	}
	return nil, false
}

// Set stores a value in the cache with the default TTL.
func Set(key string, data interface{}) {
	store.Store(key, Entry{Data: data, ExpiresAt: time.Now().Add(TTL)})
}

// Delete removes a key from the cache.
func Delete(key string) { store.Delete(key) }

// InvalidateUserStats clears the overall stats cache for a specific user.
func InvalidateUserStats(userID uint) {
	Delete(fmt.Sprintf("overall_stats:%d", userID))
}

// InvalidateModuleStats clears both overall and module-specific stats caches for a user.
func InvalidateModuleStats(moduleID, userID uint) {
	Delete(fmt.Sprintf("overall_stats:%d", userID))
	Delete(fmt.Sprintf("module_stats:%d:%d", moduleID, userID))
	Delete(fmt.Sprintf("dashboard_stats:%d", userID))
}

// InvalidateAll clears the entire cache. Used when admin modifies questions/exams/modules.
func InvalidateAll() {
	store.Range(func(key, _ interface{}) bool {
		store.Delete(key)
		return true
	})
}
