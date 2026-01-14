/**
 * Client-side data cache for Arabica
 * Caches beans, roasters, grinders, and brewers in localStorage
 * to reduce PDS round-trips on page loads.
 */

const CACHE_KEY = "arabica_data_cache";
const CACHE_VERSION = 1;
const CACHE_TTL_MS = 30 * 1000; // 30 seconds (shorter for multi-device sync)
const REFRESH_INTERVAL_MS = 30 * 1000; // 30 seconds

// Module state
let refreshTimer = null;
let isRefreshing = false;
let listeners = [];

/**
 * Get the current cache from localStorage
 */
function getCache() {
  try {
    const raw = localStorage.getItem(CACHE_KEY);
    if (!raw) return null;

    const cache = JSON.parse(raw);

    // Check version
    if (cache.version !== CACHE_VERSION) {
      localStorage.removeItem(CACHE_KEY);
      return null;
    }

    return cache;
  } catch (e) {
    console.warn("Failed to read cache:", e);
    localStorage.removeItem(CACHE_KEY);
    return null;
  }
}

/**
 * Save data to the cache
 */
function setCache(data) {
  try {
    const cache = {
      version: CACHE_VERSION,
      timestamp: Date.now(),
      data: data,
    };
    localStorage.setItem(CACHE_KEY, JSON.stringify(cache));
  } catch (e) {
    console.warn("Failed to write cache:", e);
  }
}

/**
 * Check if cache is valid (exists and not expired)
 */
function isCacheValid() {
  const cache = getCache();
  if (!cache) return false;

  const age = Date.now() - cache.timestamp;
  return age < CACHE_TTL_MS;
}

/**
 * Get cached data if available and valid
 */
function getCachedData() {
  const cache = getCache();
  if (!cache) return null;

  // Return data even if expired - caller can decide to refresh
  return cache.data;
}

/**
 * Fetch fresh data from the API
 */
async function fetchFreshData() {
  const response = await fetch("/api/data", {
    credentials: "same-origin",
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch data: ${response.status}`);
  }

  return await response.json();
}

/**
 * Refresh the cache from the API
 * Returns the fresh data
 */
async function refreshCache() {
  if (isRefreshing) {
    // Wait for existing refresh to complete
    return new Promise((resolve) => {
      const checkInterval = setInterval(() => {
        if (!isRefreshing) {
          clearInterval(checkInterval);
          resolve(getCachedData());
        }
      }, 100);
    });
  }

  isRefreshing = true;
  try {
    const data = await fetchFreshData();
    setCache(data);
    notifyListeners(data);
    return data;
  } finally {
    isRefreshing = false;
  }
}

/**
 * Get data - returns cached if valid, otherwise fetches fresh
 * @param {boolean} forceRefresh - Force a refresh even if cache is valid
 */
async function getData(forceRefresh = false) {
  if (!forceRefresh && isCacheValid()) {
    return getCachedData();
  }

  // Try to get cached data while refreshing
  const cached = getCachedData();

  try {
    return await refreshCache();
  } catch (e) {
    console.warn("Failed to refresh cache:", e);
    // Return stale data if available
    if (cached) {
      return cached;
    }
    throw e;
  }
}

/**
 * Invalidate the cache (call after CRUD operations)
 */
function invalidateCache() {
  localStorage.removeItem(CACHE_KEY);
}

/**
 * Invalidate and immediately refresh the cache
 */
async function invalidateAndRefresh() {
  invalidateCache();
  return await refreshCache();
}

/**
 * Register a listener for cache updates
 * @param {function} callback - Called with new data when cache is refreshed
 */
function addListener(callback) {
  listeners.push(callback);
}

/**
 * Remove a listener
 */
function removeListener(callback) {
  listeners = listeners.filter((l) => l !== callback);
}

/**
 * Notify all listeners of new data
 */
function notifyListeners(data) {
  listeners.forEach((callback) => {
    try {
      callback(data);
    } catch (e) {
      console.warn("Cache listener error:", e);
    }
  });
}

/**
 * Start periodic background refresh
 */
function startPeriodicRefresh() {
  if (refreshTimer) return;

  refreshTimer = setInterval(async () => {
    try {
      await refreshCache();
    } catch (e) {
      console.warn("Periodic refresh failed:", e);
    }
  }, REFRESH_INTERVAL_MS);
}

/**
 * Stop periodic background refresh
 */
function stopPeriodicRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
}

/**
 * Initialize the cache - call on page load
 * Preloads data if not cached, starts periodic refresh
 */
async function init() {
  // Start periodic refresh
  startPeriodicRefresh();

  // Preload if cache is empty or expired
  if (!isCacheValid()) {
    try {
      await refreshCache();
    } catch (e) {
      console.warn("Initial cache load failed:", e);
    }
  }

  // Refresh when user returns to tab/app (handles multi-device sync)
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible" && !isCacheValid()) {
      refreshCache().catch((e) =>
        console.warn("Visibility refresh failed:", e),
      );
    }
  });

  // For iOS PWA: refresh on focus
  window.addEventListener("focus", () => {
    if (!isCacheValid()) {
      refreshCache().catch((e) => console.warn("Focus refresh failed:", e));
    }
  });

  // Refresh on page show (back button, bfcache restore)
  window.addEventListener("pageshow", (event) => {
    if (event.persisted && !isCacheValid()) {
      refreshCache().catch((e) => console.warn("Pageshow refresh failed:", e));
    }
  });
}

/**
 * Preload cache - useful to call after login
 */
async function preload() {
  return await refreshCache();
}

// Export as global for use in other scripts
window.ArabicaCache = {
  getData,
  getCachedData,
  refreshCache,
  invalidateCache,
  invalidateAndRefresh,
  addListener,
  removeListener,
  startPeriodicRefresh,
  stopPeriodicRefresh,
  init,
  preload,
  isCacheValid,
};
