package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	goRedis "github.com/go-redis/redis/v8"
)

// redisClientPool manages shared Redis client connections
type redisClientPool struct {
	clients map[string]goRedis.UniversalClient
	mu      sync.RWMutex
}

var (
	pool     = &redisClientPool{clients: make(map[string]goRedis.UniversalClient)}
	poolOnce sync.Once
)

// getRedisClient returns a shared Redis client for the given configuration
// Multiple Cache[T] instances with the same config will share the same connection pool
func getRedisClient(addresses []string, clientConfig map[string]string, cluster bool) goRedis.UniversalClient {
	// Create a hash of the configuration to use as cache key
	configHash := hashConfig(addresses, cluster)

	pool.mu.RLock()
	client, exists := pool.clients[configHash]
	pool.mu.RUnlock()

	if exists {
		return client
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := pool.clients[configHash]; exists {
		return client
	}

	// Create new client
	var newClient goRedis.UniversalClient

	if cluster {
		newClient = goRedis.NewClusterClient(&goRedis.ClusterOptions{
			Addrs: addresses,
		})
	} else if len(addresses) > 1 {
		// Use ring for multiple addresses
		addrs := make(map[string]string)
		for i, addr := range addresses {
			addrs[string(rune('a'+i))] = addr
		}
		newClient = goRedis.NewRing(&goRedis.RingOptions{
			Addrs: addrs,
		})
	} else {
		// Single client
		newClient = goRedis.NewClient(&goRedis.Options{
			Addr: addresses[0],
		})
	}

	// Apply client configuration
	if clientConfig != nil {
		for key, value := range clientConfig {
			newClient.ConfigSet(context.Background(), key, value)
		}
	}

	pool.clients[configHash] = newClient
	return newClient
}

// hashConfig creates a unique hash for a Redis configuration
func hashConfig(addresses []string, cluster bool) string {
	h := sha256.New()
	for _, addr := range addresses {
		h.Write([]byte(addr))
	}
	if cluster {
		h.Write([]byte("cluster"))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
