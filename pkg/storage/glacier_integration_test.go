package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aldor007/stow"
	"github.com/stretchr/testify/assert"
)

// mockRestorableItem implements stow.Item and stow.Restorable for testing
type mockRestorableItem struct {
	stow.Item
	restoreCalled       bool
	restoreDays         int
	restoreTier         string
	restoreError        error
	restoreStatusResult bool
	restoreStatusExpiry *time.Time
	restoreStatusError  error
}

func (m *mockRestorableItem) Restore(ctx context.Context, days int, tier string) error {
	m.restoreCalled = true
	m.restoreDays = days
	m.restoreTier = tier
	return m.restoreError
}

func (m *mockRestorableItem) RestoreStatus(ctx context.Context) (bool, *time.Time, error) {
	return m.restoreStatusResult, m.restoreStatusExpiry, m.restoreStatusError
}

// TestIsGlacierError_WithRealProductionError verifies parsing of actual production error
func TestIsGlacierError_WithRealProductionError(t *testing.T) {
	t.Parallel()

	// Actual error from production logs (from user's issue report)
	productionError := errors.New(`<?xml version="1.0" encoding="UTF-8"?>
<Error><Code>InvalidObjectState</Code><Message>The operation is not valid for the object's storage class</Message><StorageClass>GLACIER</StorageClass><RequestId>PWTGRHAD7020T8XH</RequestId><HostId>2ysuOi8OblRjeQdP9vw76BFawNN/6eTLcpm+Svz36U0DsPH2lbB4SCpewGL5xBBZSK8dp+hElX8=</HostId></Error>`)

	isGlacier, storageClass := isGlacierError(productionError)

	assert.True(t, isGlacier, "should detect production GLACIER error")
	assert.Equal(t, "GLACIER", storageClass, "should extract GLACIER storage class")
}

// TestIsGlacierError_ErrorTypeSpecificity verifies only GLACIER errors are detected
func TestIsGlacierError_ErrorTypeSpecificity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		err                  error
		shouldBeGlacier      bool
		expectedStorageClass string
		description          string
	}{
		{
			name:                 "should detect GLACIER",
			err:                  errors.New("InvalidObjectState: GLACIER"),
			shouldBeGlacier:      true,
			expectedStorageClass: "GLACIER",
			description:          "Basic GLACIER error",
		},
		{
			name:                 "should detect DEEP_ARCHIVE",
			err:                  errors.New("InvalidObjectState: DEEP_ARCHIVE"),
			shouldBeGlacier:      true,
			expectedStorageClass: "DEEP_ARCHIVE",
			description:          "Basic DEEP_ARCHIVE error",
		},
		{
			name:                 "should NOT detect InvalidObjectState without GLACIER",
			err:                  errors.New("InvalidObjectState: Object is locked by legal hold"),
			shouldBeGlacier:      false,
			expectedStorageClass: "",
			description:          "InvalidObjectState for other reasons",
		},
		{
			name:                 "should NOT detect AccessDenied with GLACIER in message",
			err:                  errors.New("AccessDenied: No permission to access GLACIER objects"),
			shouldBeGlacier:      false,
			expectedStorageClass: "",
			description:          "Different error code even with GLACIER keyword",
		},
		{
			name:                 "should NOT detect GLACIER without InvalidObjectState",
			err:                  errors.New("MethodNotAllowed: GLACIER restore not allowed"),
			shouldBeGlacier:      false,
			expectedStorageClass: "",
			description:          "GLACIER keyword without InvalidObjectState code",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isGlacier, storageClass := isGlacierError(tt.err)

			assert.Equal(t, tt.shouldBeGlacier, isGlacier,
				"%s: isGlacier mismatch", tt.description)
			assert.Equal(t, tt.expectedStorageClass, storageClass,
				"%s: storageClass mismatch", tt.description)
		})
	}
}

// TestMockRestorableItem_InterfaceCompliance verifies mock implements interfaces
func TestMockRestorableItem_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	mock := &mockRestorableItem{}

	// Verify it implements stow.Restorable
	_, ok := interface{}(mock).(stow.Restorable)
	assert.True(t, ok, "mockRestorableItem should implement stow.Restorable")
}

// TestMockRestorableItem_RestoreMethod verifies mock restore behavior
func TestMockRestorableItem_RestoreMethod(t *testing.T) {
	t.Parallel()

	mock := &mockRestorableItem{
		restoreError: nil,
	}

	ctx := context.Background()
	err := mock.Restore(ctx, 7, "Standard")

	assert.NoError(t, err, "restore should succeed")
	assert.True(t, mock.restoreCalled, "restoreCalled should be true")
	assert.Equal(t, 7, mock.restoreDays, "should track days parameter")
	assert.Equal(t, "Standard", mock.restoreTier, "should track tier parameter")
}

// TestMockRestorableItem_RestoreStatus verifies mock status behavior
func TestMockRestorableItem_RestoreStatus(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(24 * time.Hour)
	mock := &mockRestorableItem{
		restoreStatusResult: true,
		restoreStatusExpiry: &expiry,
		restoreStatusError:  nil,
	}

	ctx := context.Background()
	inProgress, expiryTime, err := mock.RestoreStatus(ctx)

	assert.NoError(t, err, "status check should succeed")
	assert.True(t, inProgress, "should return in progress status")
	assert.NotNil(t, expiryTime, "should have expiry time")
	assert.Equal(t, expiry, *expiryTime, "expiry time should match")
}
