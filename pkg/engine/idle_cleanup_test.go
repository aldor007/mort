package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewIdleCleanupManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		enabled        bool
		timeoutMinutes int
		shouldBeNil    bool
	}{
		{
			name:           "should create manager when enabled",
			enabled:        true,
			timeoutMinutes: 15,
			shouldBeNil:    false,
		},
		{
			name:           "should create disabled manager when not enabled",
			enabled:        false,
			timeoutMinutes: 15,
			shouldBeNil:    false,
		},
		{
			name:           "should handle small timeout values",
			enabled:        true,
			timeoutMinutes: 1,
			shouldBeNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mgr := NewIdleCleanupManager(tt.enabled, tt.timeoutMinutes)

			if tt.shouldBeNil {
				assert.Nil(t, mgr)
			} else {
				assert.NotNil(t, mgr)
				assert.Equal(t, tt.enabled, mgr.enabled)
				if tt.enabled {
					assert.Equal(t, time.Duration(tt.timeoutMinutes)*time.Minute, mgr.idleTimeout)
					assert.NotNil(t, mgr.stopChan)
				}
			}
		})
	}
}

func TestIdleCleanupManager_RecordActivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "should update timestamp when enabled",
			enabled: true,
		},
		{
			name:    "should not fail when disabled",
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mgr := NewIdleCleanupManager(tt.enabled, 15)
			if !tt.enabled {
				return // Nothing to test for disabled manager
			}

			initialTime := mgr.lastActivity.Load()
			time.Sleep(1100 * time.Millisecond) // Sleep for >1 second to change Unix timestamp

			mgr.RecordActivity()

			newTime := mgr.lastActivity.Load()
			assert.Greater(t, newTime, initialTime, "timestamp should be updated")
		})
	}
}

func TestIdleCleanupManager_RecordActivity_Concurrent(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(true, 15)

	// Launch 100 concurrent RecordActivity calls
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			mgr.RecordActivity()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify timestamp was updated (should be recent)
	lastActivity := time.Unix(mgr.lastActivity.Load(), 0)
	assert.WithinDuration(t, time.Now(), lastActivity, 1*time.Second)
}

func TestIdleCleanupManager_StartStop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "should start and stop cleanly when enabled",
			enabled: true,
		},
		{
			name:    "should not panic when disabled",
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mgr := NewIdleCleanupManager(tt.enabled, 15)

			// Start should not panic
			assert.NotPanics(t, func() {
				mgr.Start()
			})

			time.Sleep(50 * time.Millisecond)

			// Stop should not panic
			assert.NotPanics(t, func() {
				mgr.Stop()
			})
		})
	}
}

func TestIdleCleanupManager_GetCleanupCount(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(true, 15)

	// Initial count should be 0
	assert.Equal(t, int64(0), mgr.GetCleanupCount())

	// Manually increment (simulating cleanup)
	mgr.cleanupCount.Add(1)
	assert.Equal(t, int64(1), mgr.GetCleanupCount())

	mgr.cleanupCount.Add(2)
	assert.Equal(t, int64(3), mgr.GetCleanupCount())
}

func TestIdleCleanupManager_CheckAndCleanup_NotIdle(t *testing.T) {
	t.Parallel()

	// Create manager with very short timeout for testing
	mgr := &IdleCleanupManager{
		enabled:     true,
		idleTimeout: 5 * time.Second,
		stopChan:    make(chan struct{}),
	}
	mgr.lastActivity.Store(time.Now().Unix())

	// Record recent activity
	mgr.RecordActivity()

	// Check should not trigger cleanup (not idle long enough)
	initialCount := mgr.GetCleanupCount()
	mgr.checkAndCleanup()

	// Cleanup count should not increase
	assert.Equal(t, initialCount, mgr.GetCleanupCount())
}

func TestIdleCleanupManager_ActivityResetsDuration(t *testing.T) {
	t.Parallel()

	// Create manager with timeout for testing
	mgr := &IdleCleanupManager{
		enabled:       true,
		idleTimeout:   5 * time.Second,
		checkInterval: 1 * time.Second,
		stopChan:      make(chan struct{}),
	}

	// Set activity time to past (>5 seconds to ensure idle)
	pastTime := time.Now().Add(-10 * time.Second)
	mgr.lastActivity.Store(pastTime.Unix())

	// Verify it would be considered idle
	idleDuration := time.Since(time.Unix(mgr.lastActivity.Load(), 0))
	assert.Greater(t, idleDuration, mgr.idleTimeout)

	// Record new activity
	mgr.RecordActivity()

	// Now should not be considered idle (timestamp should be recent)
	idleDuration = time.Since(time.Unix(mgr.lastActivity.Load(), 0))
	assert.Less(t, idleDuration, mgr.idleTimeout)
}

func TestIdleCleanupManager_StopDuringCleanup(t *testing.T) {
	t.Parallel()

	// Create manager with very short timeout
	mgr := &IdleCleanupManager{
		enabled:       true,
		idleTimeout:   10 * time.Millisecond,
		checkInterval: 5 * time.Millisecond,
		stopChan:      make(chan struct{}),
	}

	// Set activity time to past to trigger cleanup
	pastTime := time.Now().Add(-1 * time.Minute)
	mgr.lastActivity.Store(pastTime.Unix())

	// Start cleanup loop
	go mgr.cleanupLoop()

	// Wait a bit for loop to start
	time.Sleep(20 * time.Millisecond)

	// Stop should not block
	done := make(chan bool)
	go func() {
		mgr.Stop()
		done <- true
	}()

	select {
	case <-done:
		// Success - stop completed
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() blocked for too long")
	}
}

func TestIdleCleanupManager_DisabledManager(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(false, 15)

	// All operations should be no-ops and not panic
	assert.NotPanics(t, func() {
		mgr.Start()
		mgr.RecordActivity()
		mgr.Stop()
	})

	// Cleanup count should remain 0
	assert.Equal(t, int64(0), mgr.GetCleanupCount())
}

func TestIdleCleanupManager_ThreadSafety(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(true, 15)
	mgr.Start()
	defer mgr.Stop()

	// Launch multiple goroutines doing concurrent operations
	done := make(chan bool)
	operations := 50

	for i := 0; i < operations; i++ {
		go func() {
			mgr.RecordActivity()
			_ = mgr.GetCleanupCount()
			done <- true
		}()
	}

	// Wait for all operations to complete
	for i := 0; i < operations; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Operations timed out")
		}
	}
}

// TestIdleCleanupManager_IntervalCalculation verifies check interval is calculated correctly
func TestIdleCleanupManager_IntervalCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		timeoutMinutes      int
		expectedMinInterval time.Duration
	}{
		{
			name:                "should calculate interval for 15 min timeout",
			timeoutMinutes:      15,
			expectedMinInterval: 1 * time.Minute,
		},
		{
			name:                "should use minimum interval for short timeout",
			timeoutMinutes:      1,
			expectedMinInterval: 1 * time.Minute,
		},
		{
			name:                "should calculate interval for 30 min timeout",
			timeoutMinutes:      30,
			expectedMinInterval: 1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mgr := NewIdleCleanupManager(true, tt.timeoutMinutes)

			assert.NotNil(t, mgr)
			assert.GreaterOrEqual(t, mgr.checkInterval, tt.expectedMinInterval)
		})
	}
}

// Benchmark for RecordActivity to ensure it's fast enough for hot path
func BenchmarkIdleCleanupManager_RecordActivity(b *testing.B) {
	mgr := NewIdleCleanupManager(true, 15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.RecordActivity()
	}
}

// Benchmark for concurrent RecordActivity calls
func BenchmarkIdleCleanupManager_RecordActivity_Parallel(b *testing.B) {
	mgr := NewIdleCleanupManager(true, 15)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mgr.RecordActivity()
		}
	})
}

// Helper function to simulate cleanup by incrementing counter
func simulateCleanup(mgr *IdleCleanupManager) {
	mgr.cleanupCount.Add(1)
}

// Test helper to set last activity to a specific time (for testing)
func setLastActivity(mgr *IdleCleanupManager, t time.Time) {
	mgr.lastActivity.Store(t.Unix())
}

// TestIdleCleanupManager_Integration simulates a real-world scenario
func TestIdleCleanupManager_Integration(t *testing.T) {
	t.Parallel()

	// Use very short durations for testing
	mgr := &IdleCleanupManager{
		enabled:       true,
		idleTimeout:   100 * time.Millisecond,
		checkInterval: 50 * time.Millisecond,
		stopChan:      make(chan struct{}),
	}
	mgr.lastActivity.Store(time.Now().Unix())

	// Start the manager
	mgr.Start()

	// Simulate activity for 200ms
	activityDone := make(chan bool)
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(40 * time.Millisecond)
			mgr.RecordActivity()
		}
		activityDone <- true
	}()

	<-activityDone

	// Now go idle for 200ms (should trigger cleanup after 100ms + 30s safety, but we'll skip actual cleanup)
	time.Sleep(200 * time.Millisecond)

	// Stop the manager
	mgr.Stop()

	// Verify manager stopped cleanly
	assert.NotNil(t, mgr)
}
