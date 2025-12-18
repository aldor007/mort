package engine

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/h2non/bimg"
	"go.uber.org/zap"
)

// IdleCleanupManager manages memory cleanup during idle periods
// It tracks image processing activity and triggers libvips cache cleanup
// when the system has been idle for a configured duration
type IdleCleanupManager struct {
	enabled         bool
	idleTimeout     time.Duration
	checkInterval   time.Duration
	lastActivity    atomic.Int64 // Unix timestamp in seconds
	stopChan        chan struct{}
	cleanupCount    atomic.Int64 // Counter for cleanup operations (useful for metrics)
	activeProcesses atomic.Int32 // Number of active image processing operations
	cleanupMu       sync.Mutex   // Protects cleanup operations
}

// NewIdleCleanupManager creates a new IdleCleanupManager
// timeoutMinutes specifies how many minutes of inactivity before cleanup
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

	go m.cleanupLoop()
	monitoring.Log().Info("IdleCleanupManager started",
		zap.Duration("idleTimeout", m.idleTimeout),
		zap.Duration("checkInterval", m.checkInterval))
}

// Stop gracefully shuts down the cleanup goroutine
func (m *IdleCleanupManager) Stop() {
	if !m.enabled {
		return
	}

	close(m.stopChan)
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

// BeginProcessing increments the active process counter
// Call this before starting image processing to prevent cleanup during processing
func (m *IdleCleanupManager) BeginProcessing() {
	if !m.enabled {
		return
	}

	m.activeProcesses.Add(1)
	m.RecordActivity()
}

// EndProcessing decrements the active process counter
// Call this after image processing completes (use defer for safety)
func (m *IdleCleanupManager) EndProcessing() {
	if !m.enabled {
		return
	}

	m.activeProcesses.Add(-1)
}

// GetCleanupCount returns the number of cleanup operations performed
// Useful for monitoring and metrics
func (m *IdleCleanupManager) GetCleanupCount() int64 {
	return m.cleanupCount.Load()
}

// cleanupLoop is the background goroutine that periodically checks for idle state
func (m *IdleCleanupManager) cleanupLoop() {
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
// It is thread-safe and will only perform cleanup if no image processing is active
func (m *IdleCleanupManager) performCleanup() {
	// Lock to ensure only one cleanup at a time
	m.cleanupMu.Lock()
	defer m.cleanupMu.Unlock()

	// Check if there are active image processing operations
	activeCount := m.activeProcesses.Load()
	if activeCount > 0 {
		monitoring.Log().Debug("Skipping cleanup due to active processing",
			zap.Int32("activeProcesses", activeCount))
		return
	}

	// Get memory stats before cleanup
	memBefore := bimg.VipsMemory()

	monitoring.Log().Info("Performing libvips cache cleanup",
		zap.Int64("memoryBefore", memBefore.Memory),
		zap.Int64("memoryAllocations", memBefore.Allocations))

	// Perform cleanup - safe because no active processing
	bimg.VipsCacheDropAll()

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
