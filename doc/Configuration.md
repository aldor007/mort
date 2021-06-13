# Table of content

- [Configuration](#configuration)
  * [Server](#server)
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
        keys: # list of S3 keys (optional)
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

Server section describe configuration for HTTP server and some runtime variables

```yaml
server:
    listen: "0.0.0.0:8080" # default traffic listener
    monitoring: "" # default no monitoring ( or prometheus)
    cache: 
      type: "memory" # default or redis
      cacheSize: 50000 # limit of bytes used by memory cache.
      maxCacheItemSizeMB: 50 # max item size to cache default 5 MB
      # config for redis
      address:
        - "localhost:6379"
      clientConfig: # change redis instance config 
    requestTimeout: 70 # default request timeout in seconds
    internalListen: "0.0.0.0:8081" # default listener for debug /debug and metrics /metrics
    plugins: # list of additional plugins
        - "webp" # returns response based on accept header
```

## Response Headers

Overwrite response headers for given status code.

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

Main configuration for processing of request for storage or image processing. It should contain list of buckets.

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

Transform section describe if and what operation should be processed on image.

There are 3 kinds of transforms configuration:
#### Presets

```yaml
kind: "presets"
```
In this kind you have to define matching regexp for request path (path without bucket name). In regexp you have to add two matching groups - presetName, parent.
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

This kind of transform parse operations from query string. There is no need to provide regexp path.

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
In this kind you have to define matching regexp for request path (path without bucket name). In regexp you have to add two matching groups - transformations, parent.
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


