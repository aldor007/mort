# Table of content

- [Configuration](#configuration)
  * [Server](#server)
    + [Server Configuration Details](#server-configuration-details)
      - [Concurrent Image Processing](#concurrent-image-processing)
      - [Request Collapsing & Locking](#request-collapsing--locking)
      - [Cache Configuration](#cache-configuration)
      - [Idle Cleanup](#idle-cleanup)
  * [Response Headers](#response-headers)
  * [Buckets](#buckets)
    + [Transform](#transform)
      - [Presets](#presets)
      - [Query](#query)
      - [Presets-query](#presets-query)
    + [Storage](#storage)
      - [local-meta](#local-meta)
      - [noop](#noop)
      - [http](#http)
      - [s3](#s3)

# Configuration

Mort requires some configuration to run properly. This section aims at basic information about configuration.

Example config:

```yaml
headers: # add or overwrite response headers of given status. This field is optional
  - statusCodes: [200]
    values:
      "cache-control": "max-age=84000, public"

buckets: # list of available buckets
    demo:    # bucket name
        keys: # list of keys, to access mort over the S3 protocol (optional)
          - accessKey: "access"
            secretAccessKey: "random"
        transform: # config for transforms
            path: "\\/(?P<presetName>[a-z0-9_]+)\\/(?P<parent>[a-z0-9-\\.]+)" # regexp for transform path
            kind: "presets-query" #  type of transform or "query"
            presets: # list of presets
                small:
                    quality: 75
                    filters:
                        thumbnail:
                            width: 150
        storages:
             basic: # retrieve originals from s3
                 kind: "s3"
                 accessKey: "acc"
                 secretAccessKey: "sec"
                 region: ""
                 endpoint: "http://localhost:8080"
             transform: # and store it on disk
                 kind: "local-meta"
                 rootPath: "/var/www/domain/"
```

## Server

Server section describe configuration for the HTTP server and some runtime variables

```yaml
server:
    listen: "0.0.0.0:8080" # default traffic listener
    listens: # alternative: multiple listeners (HTTP and Unix sockets)
      - ":8084"
      - "unix:/tmp/mort.sock"
    monitoring: "prometheus" # default no monitoring, or "prometheus" for metrics
    logLevel: "info" # log level: debug, info, warn, error (default: info)
    accessLogs: true # enable HTTP access logs (default: false)

    # Performance & Concurrency
    concurrentImageProcessing: 100 # max concurrent image transformations (default: 100)
                                    # Adjust based on CPU/memory capacity
                                    # - Small servers: 20-50
                                    # - Medium servers (4-8 cores): 100-200
                                    # - Large servers (16+ cores): 200-500
    requestTimeout: 70 # request processing timeout in seconds (default: 60)
    lockTimeout: 30 # lock timeout for collapsed requests in seconds (default: 30)

    # Cache Configuration
    cache:
      type: "memory" # cache type: "memory" (default), "redis", or "redis-cluster"
      cacheSize: 50000 # memory cache size limit in bytes
      maxCacheItemSizeMB: 50 # max item size to cache in MB (default: 5)
      minUseCount: 2 # minimum access count before caching (prevents one-time requests)
      # Redis cache config
      address:
        - "localhost:6379"
      clientConfig: # optional redis client configuration
        maxRetries: "3"

    # Request Collapsing / Lock Configuration
    lock:
      type: "memory" # lock type: "memory" (default), "redis", or "redis-cluster"
      # Redis lock config (for distributed deployments)
      address:
        - "localhost:6379"
      clientConfig: # optional redis client configuration

    # Memory Management (optional)
    idleCleanup:
      enabled: true # enable automatic memory cleanup during idle periods
      idleTimeoutMin: 5 # minutes of inactivity before cleanup

    # Server Listeners
    internalListen: "0.0.0.0:8081" # listener for /debug (pprof) and /metrics (prometheus)

    # File Upload
    maxFileSize: 104857600 # max file upload size in bytes (default: 100MB)

    # Plugins
    plugins: # list of additional plugins
      webp: ~ # automatic WebP conversion based on Accept header
      compress: # response compression
        gzip:
          types: # MIME types to compress
            - text/plain
            - application/json
          level: 4 # compression level 1-9
        brotli:
          types:
            - text/plain
            - application/json
          level: 4
```

### Server Configuration Details

#### Concurrent Image Processing

The `concurrentImageProcessing` setting controls how many image transformations can run simultaneously. This is crucial for:

- **Preventing server overload** during traffic spikes
- **Memory management** - each transformation consumes memory
- **CPU utilization** - balancing parallel processing with system resources

**Recommended values:**
- **Default (100)**: Good for most servers with 4-8 CPU cores and 8-16GB RAM
- **Low (20-50)**: Suitable for:
  - Small servers or shared environments
  - Limited memory (< 4GB RAM)
  - Low CPU count (1-2 cores)
- **Medium (100-200)**: Suitable for:
  - Dedicated servers with 8-16 cores
  - 16-32GB RAM
  - Moderate to high traffic
- **High (200-500)**: Suitable for:
  - Large servers with 16+ cores
  - 32GB+ RAM
  - Very high traffic applications
- **Very High (1000+)**: Only for large distributed clusters with significant resources

When the limit is reached, requests return HTTP 503 (Service Unavailable) and clients can retry.

#### Request Collapsing & Locking

Mort includes a **request collapsing** mechanism to prevent duplicate processing:

- **Memory Lock** (default): Uses in-process locking, suitable for single-instance deployments
- **Redis Lock**: Uses distributed locking via Redis, required for multi-instance deployments

When multiple requests for the same image transformation arrive simultaneously:
1. First request acquires the lock and processes the image
2. Subsequent requests wait for the result
3. Once processing completes, all waiting requests receive the same result
4. This prevents the "thundering herd" problem

**Redis Lock Features:**
- Automatic lock refresh to prevent expiration during long operations
- Configurable lock timeout (default: 30 seconds)
- Pub/sub notification for efficient cross-instance communication
- Proper resource cleanup to prevent goroutine leaks

**Important:** Always call proper cleanup when shutting down to release Redis connections:
```go
if redisLock, ok := lock.(*lock.RedisLock); ok {
    redisLock.Close()
}
```

#### Cache Configuration

Response caching significantly improves performance by storing transformed images:

- **Memory Cache**: Fast, but limited to single instance
- **Redis Cache**: Shared across instances, suitable for distributed deployments

**Cache Strategy:**
- Only successful responses (HTTP 200) with known content length are cached
- Configurable maximum item size prevents memory exhaustion
- `minUseCount` prevents caching of rarely-accessed images

#### Idle Cleanup

The `idleCleanup` feature automatically releases memory during periods of low activity:
- Frees vips buffer caches
- Runs garbage collection
- Helps maintain stable memory usage over time

Enabled by default with a 5-minute idle timeout.

## Response Headers

Overwrite the response headers for a given status code.

```yaml
headers:
  - statusCodes: [200]
    values:
      "cache-control": "max-age=84000, public"
  - statusCodes: [404, 400]
    values:
      "cache-control": "max-age=60, public"
  - statusCodes: [500, 503]
    values:
      "cache-control": "max-age=10, public"
```

## Buckets

Main configuration for storage and image processing. It contains a list of buckets.

Example buckets config:

```yaml
buckets:
    media:
        keys: # s3 keys for this bucket useful for uploading files
          - accessKey: "acc"
            secretAccessKey: "sec"
        transform: # optional configuration for image operations
            path: "\\/(?P<presetName>[a-z0-9_]+)\\/(?P<parent>.*)"
            kind: "presets"
            parentBucket: "media"
            resultKey: "hash"
            presets:
                small:
                    quality: 75
                    filters:
                        thumbnail:
                            width: 150
        storages:
            basic:
                kind: "http"
                url: "https://i.imgur.com/<item>"
                headers:
                  "x--key": "sec"
            transform:
                kind: "local-meta"
                rootPath: "/Users/aldor/workspace/mkaciubacom/web"
                pathPrefix: "transforms"
```

### Transform

This section describes, if and what operation can be applied to an image.

There are 3 ways to determine which operation should be applied to an image:
#### Presets

```yaml
kind: "presets"
```
In this kind you have to define a matching regexp for request path (path without bucket name). In this regexp you have to add two matching groups - presetName, parent.
* presetName is a name of preset that should be performed on given image(parent).
* parent is a image original on which we are performing operations

Example usage:
```yaml
    path: "\\/(?P<presetName>[a-z0-9_]+)\\/(?P<parent>.*)"
```
It will match url http://mort/media/preset/dir/parent.jpg and
* presetName will be - preset
* parent will be - dir/parent.jpg

#### Query

```yaml
kind: "query"
```

This kind of transform determines the operations from the query string. There is no need to provide regexp path.

Example usage:

http://mort/media/dir/parent.jpg?operation=resize&width=1000


#### Presets-query

```yaml
kind: "presets-query"
```

This kind merge presets and query in one kind. It will try to match regexp for path it will not match then it try to parse query string.
Like in presets kind regexp is required.

Other options:

**parentBucket** - this key will add defined name to path of parent when parsing

**resultKey** - this key will define way of creating transform object unique key. Hash mean that key will be murmur hash from parent and transforms. When empty request path will be used.

**parentStorage** - change storage from with mort should fetch originals of image


**checkParent** - flag indicated that mort should always check if original object exists before returning transformation to client

#### Cloudinary

```yaml
kind: "cloudinary"
```
Like for "presets", you also have to define a matching regexp for request path (path without bucket name). In this regexp you have to add two matching groups - transformations and parent.
*
* transformations captures the part of path with transformation definiton in a Cloudinary format.
* parent is an image identifier stored in a Basic storage

Example usage:
```yaml
    path: "\\/(?P<transformations>[a-z0-9_]+)\\/(?P<parent>.*)"
```
Currently a set of supported transformation is limited to following:
 - c_fit
 - c_fill

Configuring cloudinary transform automatically enables upload support.

### Storage

This section define way of fetching object from storage. For fetching original object storage of name **basic** or defined in **parentStorage**, for image transformation
**transforms** storage will be used.

List of storage adapters:
* local-meta - adapter working on local file system
* noop - adapter that don't save image and always return that object doen't exists
* http - adapter that call remote storage using HTTP protocol
* s3 - adapter for Amazon S3 compatible service
* azure - adapter for Azure Blob Storage
* b2 - adapter for Black Base
* google - adapter for google storage
* oracle - adapter for oracle storage
* sftp - adapter for sftp

#### local-meta

Local filesystem storage.

Example definition:
```yaml
    kind: "local-meta"
    rootPath: "/Users/aldor/workspace/mkaciubacom/web" # required root path for objects
```

#### noop

No operations storage. That does nothing.

Example definition:
```yaml
    kind: "noop"
```

#### http

HTTP remote storage. That can only fetch objects.

Example definition
```yaml
    kind: "http"
    url: "http://remote/<container>/<item>"
    headers:
      "x-mort": 1
```

**url** - remote address, in url you should provide placeholders for bucket name (conatiner) and item path (item)

**headers** - additional request headers (optional)

#### s3

Adapter that fetch object from s3 storage.

Example definition
```yaml
    kind: "s3"
    accessKey: "a"
    secretAccessKey: "b"
    endpoint: "s3.amazonaws.com"
    region: "eu-west-1"
    bucket: "mybucket" # optional
```
**accessKey** - S3 access key

**secretAccessKey** - S3 secret access key

**endpoint** - address of S3 service

**region** = region of s3 service

**bucket** - bucket used for storage, when empty name of bucket will be used

#### azure

Example config

```yaml
    kind: "azure"
    azureAccount: "account"
    azureKey: "key-for-azure"
```

#### sftp

```yaml
    kind: "sftp"
    sftpHost: "sftp.dev"
    sftpPort: 22
    sftpUsername: "sftp"
    sftpPassword: "pass"
```

#### oracle

```yaml
  kind: "oracle"
  oracleUsername: "oracle"
  oraclePassword: "password"
```
#### b2

```yaml
    kind: "b2"
    b2Account: "aaa"
    b2ApplicationKeyId: "key"
```

#### google

```yaml
  kind: "google"
  googleConfigJson: |
    {"no-idea": "value"}
  googleProjectId: "id"
  googleScopes: "a, b"
```

