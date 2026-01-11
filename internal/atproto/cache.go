package atproto

import (
	"sync"
	"time"

	"arabica/internal/models"
)

// CacheTTL is how long cached data remains valid
// Set to 2 minutes to balance multi-device sync with PDS request load
const CacheTTL = 2 * time.Minute

// UserCache holds cached data for a single user.
// This struct is immutable once created - modifications create new instances.
type UserCache struct {
	Beans     []*models.Bean
	Roasters  []*models.Roaster
	Grinders  []*models.Grinder
	Brewers   []*models.Brewer
	Brews     []*models.Brew
	Timestamp time.Time
}

// IsValid returns true if the cache is still valid
func (c *UserCache) IsValid() bool {
	if c == nil {
		return false
	}
	return time.Since(c.Timestamp) < CacheTTL
}

// clone creates a shallow copy of the UserCache for safe modification
func (c *UserCache) clone() *UserCache {
	if c == nil {
		return &UserCache{Timestamp: time.Now()}
	}
	return &UserCache{
		Beans:     c.Beans,
		Roasters:  c.Roasters,
		Grinders:  c.Grinders,
		Brewers:   c.Brewers,
		Brews:     c.Brews,
		Timestamp: c.Timestamp,
	}
}

// SessionCache manages per-user caches with proper synchronization.
// Uses copy-on-write pattern to avoid race conditions when reading
// cache entries while other goroutines are modifying them.
type SessionCache struct {
	mu     sync.RWMutex
	caches map[string]*UserCache // keyed by session ID
}

// NewSessionCache creates a new session cache instance.
// Prefer this over global state for better testability and dependency injection.
func NewSessionCache() *SessionCache {
	return &SessionCache{
		caches: make(map[string]*UserCache),
	}
}

// Get retrieves a user's cache by session ID.
// The returned UserCache is safe to read without holding a lock.
func (sc *SessionCache) Get(sessionID string) *UserCache {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.caches[sessionID]
}

// Set stores a user's cache (replaces entirely)
func (sc *SessionCache) Set(sessionID string, cache *UserCache) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.caches[sessionID] = cache
}

// Invalidate removes a user's cache entirely
func (sc *SessionCache) Invalidate(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.caches, sessionID)
}

// SetBeans updates just the beans in the cache using copy-on-write
func (sc *SessionCache) SetBeans(sessionID string, beans []*models.Bean) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	newCache := sc.caches[sessionID].clone()
	newCache.Beans = beans
	newCache.Timestamp = time.Now()
	sc.caches[sessionID] = newCache
}

// SetRoasters updates just the roasters in the cache using copy-on-write
func (sc *SessionCache) SetRoasters(sessionID string, roasters []*models.Roaster) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	newCache := sc.caches[sessionID].clone()
	newCache.Roasters = roasters
	newCache.Timestamp = time.Now()
	sc.caches[sessionID] = newCache
}

// SetGrinders updates just the grinders in the cache using copy-on-write
func (sc *SessionCache) SetGrinders(sessionID string, grinders []*models.Grinder) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	newCache := sc.caches[sessionID].clone()
	newCache.Grinders = grinders
	newCache.Timestamp = time.Now()
	sc.caches[sessionID] = newCache
}

// SetBrewers updates just the brewers in the cache using copy-on-write
func (sc *SessionCache) SetBrewers(sessionID string, brewers []*models.Brewer) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	newCache := sc.caches[sessionID].clone()
	newCache.Brewers = brewers
	newCache.Timestamp = time.Now()
	sc.caches[sessionID] = newCache
}

// SetBrews updates just the brews in the cache using copy-on-write
func (sc *SessionCache) SetBrews(sessionID string, brews []*models.Brew) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	newCache := sc.caches[sessionID].clone()
	newCache.Brews = brews
	newCache.Timestamp = time.Now()
	sc.caches[sessionID] = newCache
}

// InvalidateBeans marks that beans need to be refreshed using copy-on-write
func (sc *SessionCache) InvalidateBeans(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if cache, ok := sc.caches[sessionID]; ok {
		newCache := cache.clone()
		newCache.Beans = nil
		sc.caches[sessionID] = newCache
	}
}

// InvalidateRoasters marks that roasters need to be refreshed using copy-on-write
func (sc *SessionCache) InvalidateRoasters(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if cache, ok := sc.caches[sessionID]; ok {
		newCache := cache.clone()
		newCache.Roasters = nil
		// Also invalidate beans since they reference roasters
		newCache.Beans = nil
		sc.caches[sessionID] = newCache
	}
}

// InvalidateGrinders marks that grinders need to be refreshed using copy-on-write
func (sc *SessionCache) InvalidateGrinders(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if cache, ok := sc.caches[sessionID]; ok {
		newCache := cache.clone()
		newCache.Grinders = nil
		sc.caches[sessionID] = newCache
	}
}

// InvalidateBrewers marks that brewers need to be refreshed using copy-on-write
func (sc *SessionCache) InvalidateBrewers(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if cache, ok := sc.caches[sessionID]; ok {
		newCache := cache.clone()
		newCache.Brewers = nil
		sc.caches[sessionID] = newCache
	}
}

// InvalidateBrews marks that brews need to be refreshed using copy-on-write
func (sc *SessionCache) InvalidateBrews(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if cache, ok := sc.caches[sessionID]; ok {
		newCache := cache.clone()
		newCache.Brews = nil
		sc.caches[sessionID] = newCache
	}
}

// Cleanup removes expired caches.
// This should be called periodically by a background goroutine.
func (sc *SessionCache) Cleanup() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()
	for sessionID, cache := range sc.caches {
		if now.Sub(cache.Timestamp) > CacheTTL*2 {
			delete(sc.caches, sessionID)
		}
	}
}

// StartCleanupRoutine starts a background goroutine that periodically cleans up
// expired cache entries. Returns a stop function to gracefully shut down.
func (sc *SessionCache) StartCleanupRoutine(interval time.Duration) (stop func()) {
	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				sc.Cleanup()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}
