package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGlacierError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		err                error
		expectIsGlacier    bool
		expectStorageClass string
	}{
		{
			name:               "should detect GLACIER InvalidObjectState",
			err:                errors.New("InvalidObjectState: The operation is not valid for the object's storage class. Storage class: GLACIER"),
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
		},
		{
			name:               "should detect DEEP_ARCHIVE InvalidObjectState",
			err:                errors.New("InvalidObjectState: Storage class DEEP_ARCHIVE requires restore"),
			expectIsGlacier:    true,
			expectStorageClass: "DEEP_ARCHIVE",
		},
		{
			name:               "should detect XML format error",
			err:                errors.New(`<?xml version="1.0"?><Error><Code>InvalidObjectState</Code><StorageClass>GLACIER</StorageClass></Error>`),
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
		},
		{
			name:               "should NOT detect regular not found error",
			err:                errors.New("NoSuchKey: The specified key does not exist"),
			expectIsGlacier:    false,
			expectStorageClass: "",
		},
		{
			name:               "should NOT detect access denied error",
			err:                errors.New("AccessDenied: Access Denied"),
			expectIsGlacier:    false,
			expectStorageClass: "",
		},
		{
			name:               "should NOT detect InvalidObjectState without GLACIER",
			err:                errors.New("InvalidObjectState: Some other reason"),
			expectIsGlacier:    false,
			expectStorageClass: "",
		},
		{
			name:               "should handle nil error",
			err:                nil,
			expectIsGlacier:    false,
			expectStorageClass: "",
		},
		{
			name:               "should handle generic error",
			err:                errors.New("connection timeout"),
			expectIsGlacier:    false,
			expectStorageClass: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isGlacier, storageClass := isGlacierError(tt.err)

			assert.Equal(t, tt.expectIsGlacier, isGlacier,
				"isGlacier mismatch for error: %v", tt.err)
			assert.Equal(t, tt.expectStorageClass, storageClass,
				"storageClass mismatch for error: %v", tt.err)
		})
	}
}

func TestIsGlacierError_StorageClassPriority(t *testing.T) {
	t.Parallel()

	// When error contains both GLACIER and DEEP_ARCHIVE,
	// DEEP_ARCHIVE should take precedence as it's more specific
	err := errors.New("InvalidObjectState: GLACIER DEEP_ARCHIVE")
	isGlacier, storageClass := isGlacierError(err)

	assert.True(t, isGlacier, "should detect as GLACIER error")
	assert.Equal(t, "DEEP_ARCHIVE", storageClass,
		"DEEP_ARCHIVE should take precedence over GLACIER")
}

// TestIsGlacierError_RealWorldErrors tests parsing of actual S3 error formats
func TestIsGlacierError_RealWorldErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		err                error
		expectIsGlacier    bool
		expectStorageClass string
		description        string
	}{
		{
			name: "should parse production XML error from logs",
			err: errors.New(`<?xml version="1.0" encoding="UTF-8"?>
<Error><Code>InvalidObjectState</Code><Message>The operation is not valid for the object's storage class</Message><StorageClass>GLACIER</StorageClass><RequestId>PWTGRHAD7020T8XH</RequestId><HostId>2ysuOi8OblRjeQdP9vw76BFawNN/6eTLcpm+Svz36U0DsPH2lbB4SCpewGL5xBBZSK8dp+hElX8=</HostId></Error>`),
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
			description:        "Real error from production logs",
		},
		{
			name:               "should parse AWS SDK v2 error format",
			err:                errors.New("operation error S3: GetObject, https response error StatusCode: 403, RequestID: ABC123, InvalidObjectState: The operation is not valid for the object's storage class. GLACIER"),
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
			description:        "AWS SDK v2 wrapped error format",
		},
		{
			name:               "should parse stow wrapped error",
			err:                errors.New("Open, getting the object: InvalidObjectState: storage class GLACIER not accessible"),
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
			description:        "Stow library wraps AWS errors",
		},
		{
			name:               "should parse DEEP_ARCHIVE in XML",
			err:                errors.New(`<Error><Code>InvalidObjectState</Code><StorageClass>DEEP_ARCHIVE</StorageClass></Error>`),
			expectIsGlacier:    true,
			expectStorageClass: "DEEP_ARCHIVE",
			description:        "XML format with DEEP_ARCHIVE",
		},
		{
			name:               "should parse GLACIER_IR (Instant Retrieval)",
			err:                errors.New("InvalidObjectState: GLACIER_IR storage class"),
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
			description:        "GLACIER_IR contains GLACIER keyword",
		},
		{
			name:               "should NOT parse InvalidObjectState for STANDARD",
			err:                errors.New("InvalidObjectState: Object is locked"),
			expectIsGlacier:    false,
			expectStorageClass: "",
			description:        "InvalidObjectState can occur for other reasons",
		},
		{
			name:               "should NOT parse InvalidObjectState for INTELLIGENT_TIERING",
			err:                errors.New("InvalidObjectState: INTELLIGENT_TIERING transition in progress"),
			expectIsGlacier:    false,
			expectStorageClass: "",
			description:        "Other storage classes can have InvalidObjectState",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isGlacier, storageClass := isGlacierError(tt.err)

			assert.Equal(t, tt.expectIsGlacier, isGlacier,
				"isGlacier mismatch for: %s (error: %v)", tt.description, tt.err)
			assert.Equal(t, tt.expectStorageClass, storageClass,
				"storageClass mismatch for: %s (error: %v)", tt.description, tt.err)
		})
	}
}

// TestIsGlacierError_ErrorMessageVariations tests different error message formats
func TestIsGlacierError_ErrorMessageVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		errorMessage       string
		expectIsGlacier    bool
		expectStorageClass string
	}{
		{
			name:               "should detect lowercase glacier",
			errorMessage:       "InvalidObjectState: glacier storage class",
			expectIsGlacier:    false, // Case-sensitive check
			expectStorageClass: "",
		},
		{
			name:               "should detect uppercase GLACIER",
			errorMessage:       "InvalidObjectState: GLACIER STORAGE CLASS",
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
		},
		{
			name:               "should detect GLACIER at end of message",
			errorMessage:       "InvalidObjectState: Object stored in GLACIER",
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
		},
		{
			name:               "should detect GLACIER at beginning",
			errorMessage:       "GLACIER: InvalidObjectState error",
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
		},
		{
			name:               "should detect with special characters",
			errorMessage:       "InvalidObjectState: storage_class=GLACIER, restore_required=true",
			expectIsGlacier:    true,
			expectStorageClass: "GLACIER",
		},
		{
			name:               "should NOT detect partial match",
			errorMessage:       "InvalidObjectState: GLACIAL speed",
			expectIsGlacier:    false,
			expectStorageClass: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := errors.New(tt.errorMessage)
			isGlacier, storageClass := isGlacierError(err)

			assert.Equal(t, tt.expectIsGlacier, isGlacier,
				"isGlacier mismatch for message: %s", tt.errorMessage)
			assert.Equal(t, tt.expectStorageClass, storageClass,
				"storageClass mismatch for message: %s", tt.errorMessage)
		})
	}
}

// TestIsGlacierError_ConcurrentParsing verifies thread safety
func TestIsGlacierError_ConcurrentParsing(t *testing.T) {
	t.Parallel()

	// Run concurrent error parsing to verify no race conditions
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			var err error
			if id%2 == 0 {
				err = errors.New("InvalidObjectState: GLACIER")
			} else {
				err = errors.New("InvalidObjectState: DEEP_ARCHIVE")
			}

			isGlacier, storageClass := isGlacierError(err)

			assert.True(t, isGlacier, "should detect GLACIER error")
			assert.NotEmpty(t, storageClass, "should have storage class")
			assert.Contains(t, []string{"GLACIER", "DEEP_ARCHIVE"}, storageClass,
				"storage class should be GLACIER or DEEP_ARCHIVE")

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
