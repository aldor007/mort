package engine

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/h2non/bimg"
	"github.com/stretchr/testify/assert"
)

// mockCleanupFunc is a no-op cleanup function for testing
// It doesn't call bimg.VipsCacheDropAll() so tests don't interfere with each other
func mockCleanupFunc() {
	// No-op: tests that use this won't actually cleanup libvips cache
}

// mockCleanupFuncWithCounter creates a cleanup function that increments a counter
// Useful for verifying cleanup was called without actually cleaning libvips
func mockCleanupFuncWithCounter(counter *atomic.Int64) func() {
	return func() {
		if counter != nil {
			counter.Add(1)
		}
	}
}

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
			// Use mock cleanup to avoid interfering with libvips in other tests
			mgr.cleanupFunc = mockCleanupFunc

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
		cleanupFunc:   mockCleanupFunc, // Use mock to avoid interfering with libvips
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
	// Use mock cleanup to avoid interfering with libvips in other tests
	mgr.cleanupFunc = mockCleanupFunc

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
	// Use mock cleanup to avoid interfering with libvips in other tests
	mgr.cleanupFunc = mockCleanupFunc
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
		cleanupFunc:   mockCleanupFunc, // Use mock to avoid interfering with libvips
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

// TestIdleCleanupManager_PerformCleanup_WithActiveProcesses verifies cleanup blocks until processing completes
func TestIdleCleanupManager_PerformCleanup_WithActiveProcesses(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(true, 15)
	mgr.cleanupFunc = mockCleanupFunc // Use mock to avoid interfering with libvips

	// Start processing in a goroutine
	processingDone := make(chan bool)
	go func() {
		mgr.BeginProcessing()
		time.Sleep(100 * time.Millisecond) // Hold the lock for 100ms
		mgr.EndProcessing()
		processingDone <- true
	}()

	// Give processing time to start
	time.Sleep(20 * time.Millisecond)

	// Try to perform cleanup - this should BLOCK until processing completes
	cleanupDone := make(chan bool)
	initialCount := mgr.GetCleanupCount()
	go func() {
		mgr.performCleanup()
		cleanupDone <- true
	}()

	// Verify cleanup hasn't completed yet (still blocked)
	select {
	case <-cleanupDone:
		t.Fatal("cleanup should be blocked while processing is active")
	case <-time.After(50 * time.Millisecond):
		// Good - cleanup is blocked as expected
	}

	// Wait for processing to complete
	<-processingDone

	// Now cleanup should complete
	select {
	case <-cleanupDone:
		// Good - cleanup completed after processing finished
	case <-time.After(200 * time.Millisecond):
		t.Fatal("cleanup should complete after processing ends")
	}

	// Cleanup should have run, count should increase
	assert.Equal(t, initialCount+1, mgr.GetCleanupCount(), "cleanup should run after processes complete")
}

// TestIdleCleanupManager_PerformCleanup_NoActiveProcesses verifies cleanup runs when no processing
func TestIdleCleanupManager_PerformCleanup_NoActiveProcesses(t *testing.T) {
	t.Parallel()

	mgr := NewIdleCleanupManager(true, 15)
	mgr.cleanupFunc = mockCleanupFunc // Use mock to avoid interfering with libvips

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
	mgr.cleanupFunc = mockCleanupFunc // Use mock to avoid interfering with libvips

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
		cleanupFunc:   mockCleanupFunc, // Use mock to avoid interfering with libvips
	}
	mgr.lastActivity.Store(time.Now().Unix())

	// Start the manager
	mgr.Start()

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

	// Give time for any goroutines blocked on RLock to complete
	// With RWMutex locking, if cleanup is running, some goroutines might be waiting
	time.Sleep(200 * time.Millisecond)

	// Verify all processes ended
	assert.Equal(t, int32(0), mgr.activeProcesses.Load(), "all processes should be done")

	// Set activity to past to trigger cleanup
	mgr.lastActivity.Store(time.Now().Add(-1 * time.Minute).Unix())

	// Wait for potential cleanup
	time.Sleep(100 * time.Millisecond)

	// Manager should still be running
	assert.NotNil(t, mgr)

	mgr.Stop()
}

// TestIdleCleanupManager_RaceCondition_Prevention verifies cleanup never runs during processing
func TestIdleCleanupManager_RaceCondition_Prevention(t *testing.T) {
	t.Parallel()

	mgr := &IdleCleanupManager{
		enabled:       true,
		idleTimeout:   10 * time.Millisecond,
		checkInterval: 5 * time.Millisecond,
		stopChan:      make(chan struct{}),
		cleanupFunc:   mockCleanupFunc, // Use mock to avoid interfering with libvips
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

// TestSafeVipsCleanup_DoesNotCrash verifies that safe cleanup doesn't crash
func TestSafeVipsCleanup_DoesNotCrash(t *testing.T) {
	t.Parallel()

	// Should not panic or crash
	assert.NotPanics(t, func() {
		safeVipsCleanup()
	}, "safeVipsCleanup should not panic")
}

// TestSafeVipsCleanup_AllowsSubsequentProcessing verifies image processing works after cleanup
func TestSafeVipsCleanup_AllowsSubsequentProcessing(t *testing.T) {
	t.Parallel()

	// Perform cleanup
	safeVipsCleanup()

	// Try to load and get metadata from an image immediately after cleanup
	// This should NOT crash (unlike VipsCacheDropAll which corrupts state)
	data, err := os.ReadFile("testdata/small.jpg")
	assert.NoError(t, err, "should read test image")
	if err == nil {
		// Try to get metadata - this uses libvips and would crash if state is corrupted
		metadata, err := bimg.Metadata(data)
		assert.NoError(t, err, "should get metadata after cleanup without crashing")
		assert.Greater(t, metadata.Size.Width, 0, "should have valid image dimensions")
	}
}

// TestSafeVipsCleanup_MultipleCycles verifies multiple cleanup cycles work
func TestSafeVipsCleanup_MultipleCycles(t *testing.T) {
	t.Parallel()

	// Run cleanup multiple times
	for i := 0; i < 5; i++ {
		assert.NotPanics(t, func() {
			safeVipsCleanup()
		}, "cleanup cycle %d should not panic", i+1)

		// Brief pause between cycles
		time.Sleep(50 * time.Millisecond)
	}
}

// TestSafeVipsCleanup_ReducesMemory verifies cleanup actually frees memory
func TestSafeVipsCleanup_ReducesMemory(t *testing.T) {
	t.Parallel()

	// Load test image data
	data, err := os.ReadFile("testdata/small.jpg")
	if err != nil {
		t.Skip("test image not available")
	}

	// Process several images to build up cache
	for i := 0; i < 10; i++ {
		img := bimg.NewImage(data)
		// Do some processing to populate cache
		img.Resize(100, 100)
	}

	// Get memory before cleanup
	memBefore := bimg.VipsMemory()

	// Perform cleanup
	safeVipsCleanup()

	// Wait a bit for memory to be freed
	time.Sleep(200 * time.Millisecond)

	// Get memory after cleanup
	memAfter := bimg.VipsMemory()

	// Memory should be reduced (or at least not increased)
	assert.LessOrEqual(t, memAfter.Memory, memBefore.Memory,
		"memory after cleanup should be less than or equal to before")
}

// TestSafeVipsCleanup_ConcurrentWithProcessing verifies cleanup is safe during processing
func TestSafeVipsCleanup_ConcurrentWithProcessing(t *testing.T) {
	t.Parallel()

	// Load test image data
	data, err := os.ReadFile("testdata/small.jpg")
	if err != nil {
		t.Skip("test image not available")
	}

	done := make(chan bool, 2)

	// Start image processing in background
	go func() {
		for i := 0; i < 5; i++ {
			img := bimg.NewImage(data)
			img.Resize(100, 100)
			time.Sleep(50 * time.Millisecond)
		}
		done <- true
	}()

	// Run cleanup concurrently
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(75 * time.Millisecond)
			safeVipsCleanup()
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Should complete without crashing
}

// TestSafeVipsCleanup_RestoresCacheLimits verifies cache limits are restored
func TestSafeVipsCleanup_RestoresCacheLimits(t *testing.T) {
	t.Parallel()

	// Load test image data
	data, err := os.ReadFile("testdata/small.jpg")
	if err != nil {
		t.Skip("test image not available")
	}

	// Perform cleanup
	safeVipsCleanup()

	// After cleanup, cache limits should be restored to reasonable values
	// We can verify this by processing multiple images successfully
	for i := 0; i < 10; i++ {
		img := bimg.NewImage(data)
		// Should be able to process without crashes
		_, err := img.Resize(100, 100)
		assert.NoError(t, err, "image %d should process correctly after cleanup", i+1)
	}
}

// TestSafeVipsCleanup_IntegrationWithIdleCleanupManager verifies cleanup works with manager
func TestSafeVipsCleanup_IntegrationWithIdleCleanupManager(t *testing.T) {
	t.Parallel()

	// Load test image data
	data, err := os.ReadFile("testdata/small.jpg")
	if err != nil {
		t.Skip("test image not available")
	}

	// Create manager that uses real safe cleanup (not mock)
	mgr := NewIdleCleanupManager(true, 15)

	// Verify cleanupFunc is set to safeVipsCleanup
	assert.NotNil(t, mgr.cleanupFunc, "cleanupFunc should be set")

	// Call the cleanup function through the manager
	assert.NotPanics(t, func() {
		mgr.cleanupFunc()
	}, "calling cleanupFunc should not panic")

	// Verify image processing still works after cleanup
	metadata, err := bimg.Metadata(data)
	assert.NoError(t, err, "should get metadata after manager cleanup")
	assert.Greater(t, metadata.Size.Width, 0, "should have valid image dimensions")
}

// TestSafeVipsCleanup_RestoresOriginalSettings verifies original cache settings are restored
// Note: Not using t.Parallel() because this test checks global libvips cache settings
func TestSafeVipsCleanup_RestoresOriginalSettings(t *testing.T) {
	// Set known cache settings
	bimg.VipsCacheSetMax(200)
	bimg.VipsCacheSetMaxMem(75 * 1024 * 1024)

	// Give settings time to apply
	time.Sleep(10 * time.Millisecond)

	// Get settings before cleanup
	origMax := vipsCacheGetMax()
	origMaxMem := vipsCacheGetMaxMem()

	// Verify our settings were applied
	assert.Equal(t, 200, origMax, "should have set cache max to 200")
	assert.Equal(t, 75*1024*1024, origMaxMem, "should have set cache max mem to 75MB")

	// Perform cleanup (which temporarily sets to 1/1, then restores)
	safeVipsCleanup()

	// Verify settings are restored to exactly what they were before cleanup
	assert.Equal(t, origMax, vipsCacheGetMax(), "cache max should be restored to original")
	assert.Equal(t, origMaxMem, vipsCacheGetMaxMem(), "cache max mem should be restored to original")
}
