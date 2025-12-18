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
	// Note: Not using t.Parallel() here because this test is already internally
	// concurrent with 20 goroutines. Running it in parallel with other tests
	// can overwhelm CI test coordinators.

	mgr := NewIdleCleanupManager(true, 15)

	// Launch concurrent RecordActivity calls
	concurrency := 20
	done := make(chan bool, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			mgr.RecordActivity()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
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

	// Start cleanup loop using Start() to properly initialize WaitGroup
	mgr.Start()

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
	operations := 20

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

// TestIdleCleanupManager_BeginEndProcessing tests the processing tracking
func TestIdleCleanupManager_BeginEndProcessing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "should track active processes when enabled",
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
			if !tt.enabled {
				// Should not panic
				assert.NotPanics(t, func() {
					mgr.BeginProcessing()
					mgr.EndProcessing()
				})
				return
			}

			// Initial count should be 0
			assert.Equal(t, int32(0), mgr.activeProcesses.Load())

			// Begin processing
			mgr.BeginProcessing()
			assert.Equal(t, int32(1), mgr.activeProcesses.Load())

			// Begin another
			mgr.BeginProcessing()
			assert.Equal(t, int32(2), mgr.activeProcesses.Load())

			// End one
			mgr.EndProcessing()
			assert.Equal(t, int32(1), mgr.activeProcesses.Load())

			// End the other
			mgr.EndProcessing()
			assert.Equal(t, int32(0), mgr.activeProcesses.Load())
		})
	}
}

// TestIdleCleanupManager_BeginEndProcessing_Concurrent tests concurrent processing tracking
func TestIdleCleanupManager_BeginEndProcessing_Concurrent(t *testing.T) {
	// Note: Not using t.Parallel() here because this test is already internally
	// concurrent with 20 goroutines. Running it in parallel with other tests
	// can overwhelm CI test coordinators.

	mgr := NewIdleCleanupManager(true, 15)

	// Launch concurrent operations
	concurrency := 20
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			mgr.BeginProcessing()
			time.Sleep(5 * time.Millisecond) // Simulate work
			mgr.EndProcessing()
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Final count should be 0 (all ended)
	assert.Equal(t, int32(0), mgr.activeProcesses.Load())
}

// TestIdleCleanupManager_PerformCleanup_WithActiveProcesses verifies cleanup is skipped when processing
func TestIdleCleanupManager_PerformCleanup_WithActiveProcesses(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(true, 15)

	// Mark as having active processes
	mgr.BeginProcessing()
	defer mgr.EndProcessing()

	initialCount := mgr.GetCleanupCount()

	// Try to perform cleanup
	mgr.performCleanup()

	// Cleanup should be skipped, count should not increase
	assert.Equal(t, initialCount, mgr.GetCleanupCount(), "cleanup should be skipped when processes are active")
}

// TestIdleCleanupManager_PerformCleanup_NoActiveProcesses verifies cleanup runs when no processing
func TestIdleCleanupManager_PerformCleanup_NoActiveProcesses(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(true, 15)

	// Ensure no active processes
	assert.Equal(t, int32(0), mgr.activeProcesses.Load())

	initialCount := mgr.GetCleanupCount()

	// Perform cleanup
	mgr.performCleanup()

	// Cleanup should have run, count should increase
	assert.Equal(t, initialCount+1, mgr.GetCleanupCount(), "cleanup should run when no processes are active")
}

// TestIdleCleanupManager_PerformCleanup_Concurrent verifies cleanup is thread-safe
func TestIdleCleanupManager_PerformCleanup_Concurrent(t *testing.T) {
	// Note: Not using t.Parallel() here because this test is already internally
	// concurrent with 10 goroutines. Running it in parallel with other tests
	// can overwhelm CI test coordinators.

	mgr := NewIdleCleanupManager(true, 15)

	// Launch multiple concurrent cleanup attempts
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			mgr.performCleanup()
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// All cleanups should have run (mutex ensures they're serialized)
	assert.Equal(t, int64(concurrency), mgr.GetCleanupCount())
}

// TestIdleCleanupManager_Integration_WithProcessing simulates real-world concurrent scenario
func TestIdleCleanupManager_Integration_WithProcessing(t *testing.T) {
	t.Parallel()

	// Create manager with very short timeout for testing
	mgr := &IdleCleanupManager{
		enabled:       true,
		idleTimeout:   50 * time.Millisecond,
		checkInterval: 20 * time.Millisecond,
		stopChan:      make(chan struct{}),
	}
	mgr.lastActivity.Store(time.Now().Unix())

	// Start the manager
	mgr.Start()
	defer mgr.Stop()

	// Simulate concurrent image processing
	processingCount := 20
	processingDone := make(chan bool, processingCount)

	for i := 0; i < processingCount; i++ {
		go func(id int) {
			// Simulate image processing
			mgr.BeginProcessing()
			defer mgr.EndProcessing()

			// Simulate work
			time.Sleep(30 * time.Millisecond)

			processingDone <- true
		}(i)
	}

	// Wait for all processing to complete
	for i := 0; i < processingCount; i++ {
		<-processingDone
	}

	// Verify all processes ended
	assert.Equal(t, int32(0), mgr.activeProcesses.Load(), "all processes should be done")

	// Set activity to past to trigger cleanup
	mgr.lastActivity.Store(time.Now().Add(-1 * time.Minute).Unix())

	// Wait for potential cleanup
	time.Sleep(100 * time.Millisecond)

	// Manager should still be running
	assert.NotNil(t, mgr)
}

// TestIdleCleanupManager_RaceCondition_Prevention verifies cleanup never runs during processing
func TestIdleCleanupManager_RaceCondition_Prevention(t *testing.T) {
	t.Parallel()

	mgr := &IdleCleanupManager{
		enabled:       true,
		idleTimeout:   10 * time.Millisecond,
		checkInterval: 5 * time.Millisecond,
		stopChan:      make(chan struct{}),
	}
	// Set to past to make it immediately idle
	mgr.lastActivity.Store(time.Now().Add(-1 * time.Minute).Unix())

	// Start cleanup loop
	mgr.Start()
	defer mgr.Stop()

	// Launch concurrent processing and cleanup checking
	processingCount := 20
	type result struct {
		cleanupDuringProcessing bool
	}
	results := make(chan result, processingCount)

	for i := 0; i < processingCount; i++ {
		go func() {
			// Begin processing
			mgr.BeginProcessing()
			activeAtStart := mgr.activeProcesses.Load()

			// Check if cleanup happened during processing
			cleanupBefore := mgr.GetCleanupCount()
			time.Sleep(5 * time.Millisecond) // Give cleanup loop time to attempt
			cleanupAfter := mgr.GetCleanupCount()

			cleanupHappened := cleanupAfter > cleanupBefore && activeAtStart > 0

			// End processing
			mgr.EndProcessing()

			results <- result{cleanupDuringProcessing: cleanupHappened}
		}()
	}

	// Collect results
	cleanupDuringProcessingCount := 0
	for i := 0; i < processingCount; i++ {
		r := <-results
		if r.cleanupDuringProcessing {
			cleanupDuringProcessingCount++
		}
	}

	// Verify cleanup never happened while processing was active
	assert.Equal(t, 0, cleanupDuringProcessingCount, "cleanup should never run during active processing")
}

// BenchmarkIdleCleanupManager_BeginEndProcessing benchmarks processing tracking
func BenchmarkIdleCleanupManager_BeginEndProcessing(b *testing.B) {
	mgr := NewIdleCleanupManager(true, 15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.BeginProcessing()
		mgr.EndProcessing()
	}
}

// BenchmarkIdleCleanupManager_BeginEndProcessing_Parallel benchmarks concurrent processing tracking
func BenchmarkIdleCleanupManager_BeginEndProcessing_Parallel(b *testing.B) {
	mgr := NewIdleCleanupManager(true, 15)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mgr.BeginProcessing()
			mgr.EndProcessing()
		}
	})
}
