package engine

/*
#cgo pkg-config: vips
#include <vips/vips.h>
*/
import "C"

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/h2non/bimg"
	"go.uber.org/zap"
)

// vipsCacheGetMax returns the current maximum number of cached operations
func vipsCacheGetMax() int {
	return int(C.vips_cache_get_max())
}

// vipsCacheGetMaxMem returns the current maximum cache memory in bytes
func vipsCacheGetMaxMem() int {
	return int(C.vips_cache_get_max_mem())
}

// vipsCacheGetMaxFiles returns the current maximum number of cached files
func vipsCacheGetMaxFiles() int {
	return int(C.vips_cache_get_max_files())
}

// IdleCleanupManager manages memory cleanup during idle periods
// It tracks image processing activity and triggers libvips cache cleanup
// when the system has been idle for a configured duration
type IdleCleanupManager struct {
	enabled         bool
	idleTimeout     time.Duration
	checkInterval   time.Duration
	lastActivity    atomic.Int64 // Unix timestamp in seconds
	stopChan        chan struct{}
	cleanupCount    atomic.Int64   // Counter for cleanup operations (useful for metrics)
	activeProcesses atomic.Int32   // Number of active image processing operations
	processingMu    sync.RWMutex   // RWMutex: RLock for image processing, Lock for cleanup (blocks all processing)
	wg              sync.WaitGroup // Tracks background goroutine to prevent leaks
	cleanupFunc     func()         // Function to call for cleanup (defaults to bimg.VipsCacheDropAll, can be mocked in tests)
}

// NewIdleCleanupManager creates a new IdleCleanupManager
// timeoutMinutes specifies how many minutes of inactivity before cleanup
// safeVipsCleanup performs a safe cache cleanup by temporarily reducing cache limits
// This causes libvips to naturally evict old entries without corrupting internal state
// Unlike VipsCacheDropAll(), this doesn't corrupt libvips internal hash tables
func safeVipsCleanup() {
	// Store original cache settings to restore them after cleanup
	origMax := vipsCacheGetMax()
	origMaxMem := vipsCacheGetMaxMem()
	origMaxFiles := vipsCacheGetMaxFiles()

	// Temporarily set very restrictive limits to force cache eviction
	// This causes libvips to naturally remove old cached operations
	bimg.VipsCacheSetMax(1)    // Allow only 1 cached operation
	bimg.VipsCacheSetMaxMem(1) // Allow only 1 byte cache memory

	// Brief pause to allow libvips to evict cache entries naturally
	time.Sleep(100 * time.Millisecond)

	// Restore original cache settings exactly as they were
	bimg.VipsCacheSetMax(origMax)
	bimg.VipsCacheSetMaxMem(origMaxMem)
	// Note: bimg doesn't expose VipsCacheSetMaxFiles, so we rely on the default
	_ = origMaxFiles // Keep for potential future use
}

func NewIdleCleanupManager(enabled bool, timeoutMinutes int) *IdleCleanupManager {
	if !enabled {
		return &IdleCleanupManager{enabled: false}
	}

	timeout := time.Duration(timeoutMinutes) * time.Minute
	checkInterval := timeout / 3 // Check every 1/3 of timeout period
	if checkInterval < 1*time.Minute {
		checkInterval = 1 * time.Minute
	}

	mgr := &IdleCleanupManager{
		enabled:       true,
		idleTimeout:   timeout,
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
		cleanupFunc:   safeVipsCleanup, // Use safe cleanup instead of VipsCacheDropAll
	}

	// Initialize with current time
	mgr.lastActivity.Store(time.Now().Unix())

	return mgr
}

// Start launches the background cleanup goroutine
func (m *IdleCleanupManager) Start() {
	if !m.enabled {
		return
	}

	m.wg.Add(1)
	go m.cleanupLoop()
	monitoring.Log().Info("IdleCleanupManager started",
		zap.Duration("idleTimeout", m.idleTimeout),
		zap.Duration("checkInterval", m.checkInterval))
}

// Stop gracefully shuts down the cleanup goroutine
// Waits for the background goroutine to finish to prevent goroutine leaks
func (m *IdleCleanupManager) Stop() {
	if !m.enabled {
		return
	}

	close(m.stopChan)
	m.wg.Wait() // Wait for cleanup goroutine to finish
	monitoring.Log().Info("IdleCleanupManager stopped")
}

// RecordActivity updates the last activity timestamp
// This is called on every image transform and must be very fast (lock-free)
func (m *IdleCleanupManager) RecordActivity() {
	if !m.enabled {
		return
	}

	m.lastActivity.Store(time.Now().Unix())
}

// BeginProcessing acquires a read lock and increments the active process counter
// Call this before starting image processing to prevent cleanup during processing
// This will BLOCK if cleanup is currently running
func (m *IdleCleanupManager) BeginProcessing() {
	if !m.enabled {
		return
	}

	m.processingMu.RLock() // Acquire read lock - blocks if cleanup is running
	m.activeProcesses.Add(1)
	m.RecordActivity()
}

// EndProcessing decrements the active process counter and releases the read lock
// Call this after image processing completes (use defer for safety)
func (m *IdleCleanupManager) EndProcessing() {
	if !m.enabled {
		return
	}

	m.activeProcesses.Add(-1)
	m.processingMu.RUnlock() // Release read lock
}

// GetCleanupCount returns the number of cleanup operations performed
// Useful for monitoring and metrics
func (m *IdleCleanupManager) GetCleanupCount() int64 {
	return m.cleanupCount.Load()
}

// cleanupLoop is the background goroutine that periodically checks for idle state
func (m *IdleCleanupManager) cleanupLoop() {
	defer m.wg.Done() // Signal completion when goroutine exits

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAndCleanup()
		case <-m.stopChan:
			return
		}
	}
}

// checkAndCleanup checks if the system is idle and performs cleanup if needed
func (m *IdleCleanupManager) checkAndCleanup() {
	lastActivityTime := time.Unix(m.lastActivity.Load(), 0)
	idleDuration := time.Since(lastActivityTime)

	// Check if we've been idle long enough
	if idleDuration < m.idleTimeout {
		return
	}

	monitoring.Log().Debug("Idle threshold reached, entering safety buffer",
		zap.Duration("idleDuration", idleDuration),
		zap.Duration("threshold", m.idleTimeout))

	// Safety buffer: wait 30 seconds and check again
	// This prevents cleanup if a request arrives during the check
	select {
	case <-time.After(30 * time.Second):
		// Check again after safety buffer
		lastActivityTime = time.Unix(m.lastActivity.Load(), 0)
		if time.Since(lastActivityTime) >= m.idleTimeout {
			m.performCleanup()
		} else {
			monitoring.Log().Debug("Activity detected during safety buffer, cleanup aborted")
		}
	case <-m.stopChan:
		return
	}
}

// performCleanup actually performs the libvips cache cleanup
// It acquires an exclusive write lock, blocking all image processing until cleanup completes
// This ensures no images are being processed during cleanup
func (m *IdleCleanupManager) performCleanup() {
	// Acquire write lock - this will:
	// 1. Wait for all current image processing (RLock holders) to complete
	// 2. Block any new image processing from starting
	m.processingMu.Lock()
	defer m.processingMu.Unlock()

	// Safety check: verify no active processes (should always be 0 due to lock)
	activeCount := m.activeProcesses.Load()
	if activeCount != 0 {
		monitoring.Log().Error("CRITICAL: Active processes detected despite holding cleanup lock!",
			zap.Int32("activeProcesses", activeCount))
		return
	}

	monitoring.Log().Info("Acquired cleanup lock, all image processing blocked",
		zap.Int32("activeProcesses", activeCount))

	// Get memory stats before cleanup
	memBefore := bimg.VipsMemory()

	monitoring.Log().Info("Performing libvips cache cleanup",
		zap.Int64("memoryBefore", memBefore.Memory),
		zap.Int64("memoryAllocations", memBefore.Allocations))

	// Perform cleanup - safe because we have exclusive lock
	// Use the injected cleanup function (defaults to bimg.VipsCacheDropAll)
	if m.cleanupFunc != nil {
		m.cleanupFunc()
	}

	// Get memory stats after cleanup
	memAfter := bimg.VipsMemory()

	// Increment cleanup counter
	m.cleanupCount.Add(1)

	// Log results
	memFreed := memBefore.Memory - memAfter.Memory
	monitoring.Log().Info("Libvips cache cleanup completed",
		zap.Int64("memoryAfter", memAfter.Memory),
		zap.Int64("memoryFreed", memFreed),
		zap.Int64("totalCleanups", m.cleanupCount.Load()))

	// Report to monitoring if available
	if memFreed > 0 {
		monitoring.Report().Inc("vips_cleanup_count")
	}
}
