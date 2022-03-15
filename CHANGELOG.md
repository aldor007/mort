# [0.15.0](https://github.com/aldor007/mort/compare/v0.14.1...v0.15.0) (2022-03-15)


### Bug Fixes

* **docker:** add missing ca-certificates ([335fc62](https://github.com/aldor007/mort/commit/335fc622c98e97207e3f71c0ecb82b1166d7e99a))
* fix cache key ([da5a722](https://github.com/aldor007/mort/commit/da5a7225c766fafe3090b8d7da8d4dce15ea75a1))
* fix docker build tag ([1964270](https://github.com/aldor007/mort/commit/196427082a7cdff93f6b0dd6d43f337708e6af16))
* fix docker build tag ([0883c35](https://github.com/aldor007/mort/commit/0883c3525277ab8f92184c110914fb7b89c8c369))
* fix resizeCropAuto when one parameter is 0 ([ffbdeea](https://github.com/aldor007/mort/commit/ffbdeeace392cc31eddf25789eeaf86ecee97c11))
* fix status code of error response puted into cache ([c4a213d](https://github.com/aldor007/mort/commit/c4a213d7a3c9cfb1eab156d84e35fd86b4792994))
* fix transform for resizeCropAuto and resize in one take ([f06af8c](https://github.com/aldor007/mort/commit/f06af8c798b53b773a35445e24d61ab3ff510496))
* **redis:** copy fileobject ([ac24a63](https://github.com/aldor007/mort/commit/ac24a6378f38f4725e3a766465e2596c0df32e16))
* **redis:** redis cluster does not support db ([dc21c1b](https://github.com/aldor007/mort/commit/dc21c1bd3f7c03d8b465b2aeaa4da95b7f0523d2))
* run docker and release after semantic release ([18bd855](https://github.com/aldor007/mort/commit/18bd855bee9a2f47ca3cc7b1ff97cb6a00cf05f6))
* run docker and release after semantic release ([c8c871a](https://github.com/aldor007/mort/commit/c8c871a0b7349ec6819ae22b56f831786eac3563))
* run docker and release after semantic release ([b980cdb](https://github.com/aldor007/mort/commit/b980cdb0c5ba65a4ceea3b47e22649f9636f80f5))
* update bimg - fix blury webp images ([f053faa](https://github.com/aldor007/mort/commit/f053faa82faf50e671525d3551d78b59a07c9708))


### Features

*  update libvips and golang ([75f6fbd](https://github.com/aldor007/mort/commit/75f6fbdc02253c96d38a82374261c1e8fef000a9))
* add user defined script for parsing URL ([ca950c1](https://github.com/aldor007/mort/commit/ca950c148aecaf72363b5f659a2b4fe72109003c))
* **docker:** update ubuntu image ([04c926b](https://github.com/aldor007/mort/commit/04c926bd48d02071fc525d42822f2fab77b64183))
* **docker:** update ubuntu image ([ed9337f](https://github.com/aldor007/mort/commit/ed9337f2c02a878ce3f57838e115990f2d5d748c))
* **docker:** update ubuntu image ([c849480](https://github.com/aldor007/mort/commit/c849480140062207d380458dd3a62837f106c0d1))
* **headers:** add flag that allow force override headers ([a76208f](https://github.com/aldor007/mort/commit/a76208f1729c944b0d2b6741f8fca135a0ff0e3f))
* **redis:** update redis client and handle redis cluster && update golang ([1b31c0b](https://github.com/aldor007/mort/commit/1b31c0b10d6a51c690144b11fb0beae94e2c8dc0))
* update tests, new params for resize ([10e28bc](https://github.com/aldor007/mort/commit/10e28bc7d3650354a43629f6577e39313921833c))
* use base image for mort image ([17f6760](https://github.com/aldor007/mort/commit/17f67603d675396bbc5efd96d4a6a7277ee687cf))

# [0.18.0](https://github.com/aldor007/mort/compare/0.17.1...0.18.0) (2022-03-14)


### Features

* use base image for mort image ([17f6760](https://github.com/aldor007/mort/commit/17f67603d675396bbc5efd96d4a6a7277ee687cf))

## [0.17.1](https://github.com/aldor007/mort/compare/0.17.0...0.17.1) (2022-03-14)


### Bug Fixes

* run docker and release after semantic release ([18bd855](https://github.com/aldor007/mort/commit/18bd855bee9a2f47ca3cc7b1ff97cb6a00cf05f6))
* run docker and release after semantic release ([c8c871a](https://github.com/aldor007/mort/commit/c8c871a0b7349ec6819ae22b56f831786eac3563))
* run docker and release after semantic release ([b980cdb](https://github.com/aldor007/mort/commit/b980cdb0c5ba65a4ceea3b47e22649f9636f80f5))

# [0.17.0](https://github.com/aldor007/mort/compare/0.16.1...0.17.0) (2022-03-14)


### Bug Fixes

* fix resizeCropAuto when one parameter is 0 ([ffbdeea](https://github.com/aldor007/mort/commit/ffbdeeace392cc31eddf25789eeaf86ecee97c11))


### Features

* add user defined script for parsing URL ([ca950c1](https://github.com/aldor007/mort/commit/ca950c148aecaf72363b5f659a2b4fe72109003c))

* 0.14.1
     * Bugfix: Fix copying headers from object for S3 
* 0.14.0
     * Feature: Allow to define headers per bucket
* 0.13.0
     * Feature: Redis response cache
     * Feature: Extract transform
     * Feature: ResizeCropAuto transform
     * Feature: handle b2 storage
* 0.12.0    
    * Feature: Add new monitoring metrics (time of image generation and count of it)
    * Feature: Do error placeholder in background (returns it faster to user)
    * Feature: Try to merge transformations before performing them
* 0.11.2
    * Bugfix: Fix compress plugin (don't compress on range or condition)
* 0.11.1
    * Bugfix: Fix compress plugin (it was returning invalid headers, Content-Encoding: gzip/br when no compression)
* 0.11.0
    * Feature: Compress plugin (gzip, brotli)
* 0.10.0
    * Feature: Introduce plugins
    * Feature: Webp plugins (if there is image/webp in accept header convert object to that format)
* 0.9.1
    * Bugfix: Fixed goroutines leak
    * Feature: Added lockTimeout to config 
* 0.9.0
    * Feature: Allow to define placeholder for error response 
* 0.8.0
    * Feature: Added support for AWS presigned url
    * Bugfix: Fixed reporting collaped reqeuest to promethues
    * Feature: Update golang to 1.10 update libvips 8.6.2 
* 0.7.0
    * Feature: Remove mutex from time monitoring 
* 0.6.1
    * Bugfix: Add mutex for time monitoring map 
* 0.6.0
    * Feature: Add prometheus reporter
    * Bugfix: Fix race condition when notifying waiting clients about response 
* 0.5.0
    * Feature: Allow to have multiple listeners
    * Bugfix: Add locking for storage and preset cache
* 0.4.2
    * Bugfix: Update stow (fixed removing dir) 
* 0.4.1
    * Bugfix: Update stow (fixed serving range response for object with local-meta adapter) 
* 0.4.0
    * Feature: Implement delete object from storage
    * Bugfix: Fixed handling AWS Auth header from aws-php-sdk
* 0.3.0
    * Feature: Handle Range and condition request
* 0.2.1
    * Bugfix: Fixed regression in calculating parent in presets parser
* 0.2.0
    * Feature: Allow register custom url parser 
    * Bugfix: Change ETag to weak ETag
    * Bugfix: Fixed race when getting object from cache
* 0.1.0
    * Feature: New kind of transformation presets-query it combines presets and query in on kind
    * Feature: Added gravity to crop transform
    * Feature: Allow to configure server properties like listen addr, cache size etc
    * Bugfix: Fixed position for watermark transform
    * Bugfix: Change collapse key from request path to object key
* 0.0.1 
    * Initial release
