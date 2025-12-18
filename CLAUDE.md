# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mort is an S3-compatible image processing server written in Go. It transforms images on-the-fly using URL-based parameters (presets or query strings) and supports multiple storage backends (S3, local, HTTP, Azure, Google Cloud, etc.). The server handles request collapsing, rate limiting, caching, S3 GLACIER object restoration, and includes an S3-compatible API for listing and uploading files.

**New Feature:** Automatic S3 GLACIER/DEEP_ARCHIVE object restore - see [docs/GLACIER_RESTORE.md](docs/GLACIER_RESTORE.md)

## Build & Development Commands

### Running Tests
```bash
# Run unit tests with race detection and formatting
make unit

# Run unit tests with benchmarks
make unit-bench

# Run integration tests (requires npm dependencies)
make integrations

# Run all tests
make tests

# Run tests in docker
make docker-tests

# Generate coverage report
make coverage
```

### Running Single Tests
```bash
# Run specific test with race detection
go test -race -run TestName ./pkg/path/to/package

# Run tests in a single package
go test -race ./pkg/cache/...

# Run with verbose output
go test -v -race -run TestName ./pkg/path/to/package
```

### Running the Server
```bash
# Run with default config
make run-server

# Run with custom config
go run cmd/mort/mort.go -config path/to/config.yml

# Run test server
make run-test-server

# Run test server with Redis
make run-test-server-redis
```

### Building & Formatting
```bash
# Format and vet code
make format

# Build binary
go build -o mort cmd/mort/mort.go
```

## Architecture

### Request Flow
1. **HTTP Request** → chi router with middleware (S3 auth, cloudinary upload interceptor)
2. **FileObject Creation** → Parses URL into a FileObject with bucket, key, transforms, and storage config
3. **Request Processor** → Core component that orchestrates the entire processing pipeline
4. **Response Cache Check** → Checks if transformed image is already cached
5. **Request Collapsing** → Multiple concurrent requests for same resource are collapsed into one
6. **Storage Layer** → Retrieves original image from configured storage backend
7. **Image Engine** → Applies transforms using libvips (via bimg)
8. **Response** → Returns transformed image and caches it

### Core Components

#### pkg/processor/processor.go
The heart of Mort. `RequestProcessor` handles all incoming requests with:
- Request collapsing using `lock.Lock` interface to prevent duplicate processing
- Rate limiting via `throttler.Throttler` to control concurrent image transformations
- Response caching to serve repeated requests quickly
- Parent checking to verify original images exist before transformation
- Timeout handling for long-running operations

#### pkg/object/file_object.go
`FileObject` represents a parsed request containing:
- `Bucket`: which bucket config to use
- `Key`: storage path for the file
- `Transforms`: list of transformations to apply
- `Storage`: which storage backend to use
- `Parent`: reference to original image (for transformed images)

URL parsing supports multiple modes:
- **presets**: predefined transformations (e.g., `/bucket/small/image.jpg`)
- **query**: query string transforms (e.g., `/bucket/image.jpg?width=100&height=100`)
- **presets-query**: combination of both
- **tengo**: custom URL parser using Tengo scripting language

#### pkg/storage/storage.go
Abstraction over multiple storage backends using the `stow` library:
- Supports: local, local-meta, s3, http, b2, google, azure, sftp, oracle
- Provides: Get, Head, Set, Delete, List operations
- Thread-safe storage client caching via `storageCache` map with RWMutex
- Handles S3-compatible API for listing objects

#### pkg/config/config.go
Singleton configuration loaded from YAML with environment variable expansion:
- Bucket configurations with transforms, storages, and access keys
- Storage configurations for different backends
- Server settings (ports, timeouts, cache settings)
- Transform rules (regex patterns, presets)

#### pkg/engine/ (image processing)
Wraps bimg (libvips bindings) to perform image transformations:
- Resize, crop, rotate, blur, watermark, format conversion
- Smart cropping using feature detection
- Quality and compression settings

#### pkg/cache/
Response caching implementations:
- `MemoryCache`: in-memory cache using ccache
- `RedisCache`: distributed cache using Redis
- Caches full HTTP responses including headers and body

#### pkg/lock/
Request collapsing implementations:
- `MemoryLock`: in-process lock using sync.Map
- `RedisLock`: distributed lock using Redis (with redislock library)
- Allows one request to process an image while others wait for the result

### Storage Path Handling
- Basic storage: retrieves original images (configured per bucket)
- Transform storage: stores processed/transformed images
- Path prefixes can be configured per storage to organize files
- S3 and B2 require trimming leading slashes from keys

## Configuration

Configuration is YAML-based with environment variable support (`${VAR_NAME}`). Key concepts:

### Bucket Structure
```yaml
buckets:
  bucket-name:
    keys:  # S3 API access keys (optional)
      - accessKey: "key"
        secretAccessKey: "secret"
    transform:
      path: "regexp pattern with named groups"
      kind: "presets|query|presets-query|tengo"
      presets:  # for preset-based transforms
        preset-name:
          quality: 80
          filters: {...}
    storages:
      basic:  # for original images
        kind: "s3|local|http|..."
      transform:  # for processed images
        kind: "local-meta"
        pathPrefix: "transforms"
```

### Storage Types
Each storage backend has specific required fields (see `pkg/config/config.go` `validateStorage`):
- `local`/`local-meta`: requires `rootPath`
- `s3`: requires `accessKey`, `secretAccessKey`, `region`, `endpoint`
- `http`: requires `url`
- `b2`, `google`, `azure`, `sftp`, `oracle`: see respective config fields

## Testing Conventions

- Use `testify/assert` for assertions
- Table-driven tests are preferred for multiple test cases
- Always run tests with `-race` flag to detect race conditions
- Use `t.Parallel()` for tests that can run concurrently
- Test files are named `*_test.go` and placed alongside source files

## Important Implementation Details

### libvips/bimg Setup
Requires libvips installed on the system. If encountering build errors with `pkg-config --cflags`, set:
```bash
export CGO_CFLAGS_ALLOW="-Xpreprocessor"
```

### Lock/Collapse Mechanism
The request collapsing prevents the "thundering herd" problem:
1. First request acquires lock for an image key
2. Subsequent requests wait on a channel for the result
3. When processing completes, the result is broadcast to all waiting requests
4. Includes timeout handling (default 30s) to prevent indefinite waits

### Response Caching Strategy
- Only caches successful responses (200 status) with known content length
- Max cache item size configurable (default 5MB)
- Cache key includes transform hash for unique identification
- Supports both memory and Redis-based caching

### Parent Checking
For transformed images, Mort can verify the original (parent) image exists:
- `CheckParent` flag controls this behavior
- Useful for S3 API where listing queries need parent validation
- Trades performance for correctness

### Timeouts
- `RequestTimeout`: overall request processing timeout (default 60s)
- `LockTimeout`: max wait time for collapsed requests (default 30s)
- Only applied to requests with transforms (large file uploads have no timeout)

## Common Development Patterns

### Adding New Storage Backend
1. Register storage kind in `storageKinds` array in `pkg/config/config.go`
2. Add config validation in `validateStorage`
3. Add stow configuration in `getClient` function in `pkg/storage/storage.go`
4. Add config fields to `config.Storage` struct

### Adding New Transform Type
1. Implement URL parser (see `pkg/object/query.go`, `preset.go` as examples)
2. Register transform kind using `config.RegisterTransformKind()`
3. Add validation in `validateTransform`
4. Parser must populate `FileObject.Transforms` and `FileObject.Key`

### Adding New Image Operation
1. Add filter implementation in `pkg/processor/plugins/` or extend `pkg/engine/`
2. Update transform parsing to support new operation
3. Filters are applied in order specified in configuration
- memorise commend with flags for tests