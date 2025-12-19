package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRedisClient_SharedConnectionPool(t *testing.T) {
	t.Parallel()

	addresses := []string{"localhost:6379"}
	clientConfig := map[string]string{}

	// Create first client
	client1 := getRedisClient(addresses, clientConfig, false)
	assert.NotNil(t, client1, "first client should be created")

	// Create second client with same config
	client2 := getRedisClient(addresses, clientConfig, false)
	assert.NotNil(t, client2, "second client should be created")

	// Should return the SAME client instance (shared pool)
	assert.Same(t, client1, client2, "clients with same config should be the same instance")
}

func TestGetRedisClient_DifferentConfigurations(t *testing.T) {
	t.Parallel()

	// Client 1: single address
	client1 := getRedisClient([]string{"localhost:6379"}, nil, false)

	// Client 2: different address
	client2 := getRedisClient([]string{"localhost:6380"}, nil, false)

	// Should be different clients (different configs)
	assert.NotSame(t, client1, client2, "clients with different addresses should be different instances")
}

func TestGetRedisClient_ClusterVsNonCluster(t *testing.T) {
	t.Parallel()

	addresses := []string{"localhost:6379"}

	// Non-cluster client
	client1 := getRedisClient(addresses, nil, false)

	// Cluster client with same address
	client2 := getRedisClient(addresses, nil, true)

	// Should be different clients (different cluster setting)
	assert.NotSame(t, client1, client2, "cluster and non-cluster clients should be different")
}

func TestHashConfig_Consistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		addresses []string
		cluster   bool
	}{
		{
			name:      "should generate consistent hash for same config",
			addresses: []string{"localhost:6379"},
			cluster:   false,
		},
		{
			name:      "should generate consistent hash for multiple addresses",
			addresses: []string{"host1:6379", "host2:6379"},
			cluster:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hashConfig(tt.addresses, tt.cluster)
			hash2 := hashConfig(tt.addresses, tt.cluster)

			assert.Equal(t, hash1, hash2, "hash should be consistent for same config")
			assert.NotEmpty(t, hash1, "hash should not be empty")
		})
	}
}

func TestHashConfig_Different(t *testing.T) {
	t.Parallel()

	hash1 := hashConfig([]string{"localhost:6379"}, false)
	hash2 := hashConfig([]string{"localhost:6380"}, false)
	hash3 := hashConfig([]string{"localhost:6379"}, true)

	assert.NotEqual(t, hash1, hash2, "different addresses should have different hashes")
	assert.NotEqual(t, hash1, hash3, "different cluster setting should have different hashes")
}
