# [0.29.0](https://github.com/aldor007/mort/compare/v0.28.0...v0.29.0) (2023-05-21)


### Features

* update aws sdk to v2 fix-v2 ([cc31c78](https://github.com/aldor007/mort/commit/cc31c78794def7169b6a7a1308dad04f1c798a16))

# [0.28.0](https://github.com/aldor007/mort/compare/v0.27.1...v0.28.0) (2023-05-21)


### Features

* update aws sdk to v2 fix ([0af0c58](https://github.com/aldor007/mort/commit/0af0c581d8201333996412106a8ce684b5d40859))

## [0.27.1](https://github.com/aldor007/mort/compare/v0.27.0...v0.27.1) (2023-05-21)


### Bug Fixes

* trigger release - test deployment ([29a2abb](https://github.com/aldor007/mort/commit/29a2abbe44566f56f3625304d42a29a34209c8a6))

# [0.27.0](https://github.com/aldor007/mort/compare/v0.26.1...v0.27.0) (2023-05-21)


### Features

* update aws sdk to v2 - stow ([f5eccf3](https://github.com/aldor007/mort/commit/f5eccf3582922e0eb068a58f89f225c027384c53))

## [0.26.1](https://github.com/aldor007/mort/compare/v0.26.0...v0.26.1) (2022-10-25)


### Bug Fixes

* config name for minUseCount ([e81c1de](https://github.com/aldor007/mort/commit/e81c1de6bb0dfbf9f0e9eddaedc5281c79c2b73f))
* remove counter if usage count is grater ([20c6299](https://github.com/aldor007/mort/commit/20c6299ccb37ce00e26b8e230742a15530bb6362))

# [0.26.0](https://github.com/aldor007/mort/compare/v0.25.0...v0.26.0) (2022-10-25)


### Features

* redis min use counter ([4a52e11](https://github.com/aldor007/mort/commit/4a52e1179a19084b03a6d829c85609494d88e0bf))

# [0.25.0](https://github.com/aldor007/mort/compare/v0.24.2...v0.25.0) (2022-10-24)


### Features

* update ccache dep ([891ee36](https://github.com/aldor007/mort/commit/891ee36b6d6834a7ec57f4b9ab43125242cd3f29))

## [0.24.2](https://github.com/aldor007/mort/compare/v0.24.1...v0.24.2) (2022-10-24)


### Bug Fixes

* memory lock panic [#49](https://github.com/aldor007/mort/issues/49) ([4b4b5f9](https://github.com/aldor007/mort/commit/4b4b5f9170a4e809dcebefd0038a22d0098211d2))

## [0.24.1](https://github.com/aldor007/mort/compare/v0.24.0...v0.24.1) (2022-10-24)


### Bug Fixes

* memory lock panic [#49](https://github.com/aldor007/mort/issues/49) ([e4bba26](https://github.com/aldor007/mort/commit/e4bba26bb9f7f166a582f5223380475835fbdcd8))

# [0.24.0](https://github.com/aldor007/mort/compare/v0.23.7...v0.24.0) (2022-10-21)


### Features

* parent image cache ([b5a80e4](https://github.com/aldor007/mort/commit/b5a80e4a0a2900f11b63c8345da501b07e7fa5ed))

## [0.23.7](https://github.com/aldor007/mort/compare/v0.23.6...v0.23.7) (2022-10-20)


### Bug Fixes

* redis cache tests ([3e81a21](https://github.com/aldor007/mort/commit/3e81a2179860e50180a82d41e716781a5ebf954a))

## [0.23.6](https://github.com/aldor007/mort/compare/v0.23.5...v0.23.6) (2022-10-20)


### Bug Fixes

* redis lock improvments ([e426e61](https://github.com/aldor007/mort/commit/e426e61e38c261379a17e3ad03e600742694f8bf))
* redis lock improvments && update CI tests ([e1da289](https://github.com/aldor007/mort/commit/e1da289921a0f54cc7faac9ac7ec764dc9b49261))

## [0.23.5](https://github.com/aldor007/mort/compare/v0.23.4...v0.23.5) (2022-10-20)


### Bug Fixes

* fix redis lock ([b1bbb6d](https://github.com/aldor007/mort/commit/b1bbb6d658efe58010260bb7b51ab1c03a5ab4f1))

## [0.23.4](https://github.com/aldor007/mort/compare/v0.23.3...v0.23.4) (2022-10-19)


### Bug Fixes

* gorouting leak on redis lock ([453bc3c](https://github.com/aldor007/mort/commit/453bc3c228f1bd3cc421e774c9cbc9309a43620a))
* gorouting leak on redis lock - fmt not used ([2ad402f](https://github.com/aldor007/mort/commit/2ad402fc6f447af50484d8718e1ef7c446d71f6f))

## [0.23.3](https://github.com/aldor007/mort/compare/v0.23.2...v0.23.3) (2022-08-19)


### Bug Fixes

* preSign redirect ([c7cb52f](https://github.com/aldor007/mort/commit/c7cb52f4ca8296eb9b3329a90e34195f3680afa2))

## [0.23.2](https://github.com/aldor007/mort/compare/v0.23.1...v0.23.2) (2022-07-29)


### Bug Fixes

* close response when doing preSign request ([a4ceef1](https://github.com/aldor007/mort/commit/a4ceef1f814c057474e313194120434549849c04))

## [0.23.1](https://github.com/aldor007/mort/compare/v0.23.0...v0.23.1) (2022-07-29)


### Bug Fixes

* use GET in preSign ([272e0da](https://github.com/aldor007/mort/commit/272e0daf5f959f55f44da5ca767ae4ae255c2417))

# [0.23.0](https://github.com/aldor007/mort/compare/v0.22.0...v0.23.0) (2022-07-28)


### Bug Fixes

* fix ci ([313d7c8](https://github.com/aldor007/mort/commit/313d7c8c5b838a32901ec12220d97999cffa1678))


### Features

* update docker ([4d08c49](https://github.com/aldor007/mort/commit/4d08c497647f1475d2f60b51336f1f3496e2b85b))

# [0.22.0](https://github.com/aldor007/mort/compare/v0.21.5...v0.22.0) (2022-07-28)


### Features

* introduce maxFileSize server config that uses redirect for big files, use go1.18 ([97ae2ea](https://github.com/aldor007/mort/commit/97ae2eaa326b0087daaa388a1ce9a40739737a7f))
* preSign url ([a5c84d0](https://github.com/aldor007/mort/commit/a5c84d002e28719bdda7476c45908f59860d7f38))

## [0.21.5](https://github.com/aldor007/mort/compare/v0.21.4...v0.21.5) (2022-03-19)


### Bug Fixes

* **compress:** fix response for compression plugin and add tests ([831a816](https://github.com/aldor007/mort/commit/831a8166d84a3614d5d0602ab32235e26f54686d))

## [0.21.4](https://github.com/aldor007/mort/compare/v0.21.3...v0.21.4) (2022-03-17)


### Bug Fixes

* do not write to cancel channel on error placeholder generation - goroutine leak ([80bad48](https://github.com/aldor007/mort/commit/80bad48f5f303fc48adf948f294b196399505056))

## [0.21.3](https://github.com/aldor007/mort/compare/v0.21.2...v0.21.3) (2022-03-17)


### Bug Fixes

* **response:** return content-length in reponse ([a9bf75a](https://github.com/aldor007/mort/commit/a9bf75aac059c86b4fbd9d4b48bd49bca11a523a))

## [0.21.2](https://github.com/aldor007/mort/compare/v0.21.1...v0.21.2) (2022-03-17)


### Bug Fixes

* **timeout:** increase server read and write timeout - issue with big file download ([d6fec5f](https://github.com/aldor007/mort/commit/d6fec5f7842f3da140574206e2a19ae2157c1e11))

## [0.21.1](https://github.com/aldor007/mort/compare/v0.21.0...v0.21.1) (2022-03-17)


### Bug Fixes

* remove processTimeout for non image request ([8089314](https://github.com/aldor007/mort/commit/8089314ded17682919573680b992c6f9b05a3df5))

# [0.21.0](https://github.com/aldor007/mort/compare/v0.20.0...v0.21.0) (2022-03-17)


### Features

* **logs:** add access log ([96c6175](https://github.com/aldor007/mort/commit/96c6175510149867fcd2c2c2d2cbeabb9bc46d19))

# [0.20.0](https://github.com/aldor007/mort/compare/v0.19.2...v0.20.0) (2022-03-16)


### Features

* **storage:** handle more storage, update stow ([e55ac87](https://github.com/aldor007/mort/commit/e55ac87395a8aec8f60158134a79aac3d70c19cb))

## [0.19.2](https://github.com/aldor007/mort/compare/v0.19.1...v0.19.2) (2022-03-16)


### Bug Fixes

* **lock:** fix redis lock config key name ([a96097d](https://github.com/aldor007/mort/commit/a96097d5fd5367ff640fb5e930632cd87f8d3abd))

## [0.19.1](https://github.com/aldor007/mort/compare/v0.19.0...v0.19.1) (2022-03-16)


### Bug Fixes

* **logs:** add logs about cache and lock strategy ([59e7965](https://github.com/aldor007/mort/commit/59e7965a737180e243ac04be350f881defae3713))

# [0.19.0](https://github.com/aldor007/mort/compare/v0.18.2...v0.19.0) (2022-03-16)


### Features

* **lock:** use redis pubsub for redislock strategy and refactor logs ([b54e597](https://github.com/aldor007/mort/commit/b54e597b3fe20ea122ef5a8405ab61c1aa344461))

## [0.18.2](https://github.com/aldor007/mort/compare/v0.18.1...v0.18.2) (2022-03-15)


### Bug Fixes

* **redis:** allow to configure redis lock timeoutk ([af97dff](https://github.com/aldor007/mort/commit/af97dffda5c499de11ac48239ff5427493fbb806))

## [0.18.1](https://github.com/aldor007/mort/compare/v0.18.0...v0.18.1) (2022-03-15)


### Bug Fixes

* **redis:** fix handling unable to acquire redis lock ([7afdf6f](https://github.com/aldor007/mort/commit/7afdf6f089d44e01229d8f2913eef8320e4a8263))

# [0.18.0](https://github.com/aldor007/mort/compare/v0.17.0...v0.18.0) (2022-03-15)


### Features

* **redis:** add redis lock strategy ([c26fd20](https://github.com/aldor007/mort/commit/c26fd2035a3b5c1d6f7afb9717ad8bd25349fd23))

# [0.17.0](https://github.com/aldor007/mort/compare/v0.16.3...v0.17.0) (2022-03-15)


### Features

* **tengo:** handling errors, preset example ([39cb9d5](https://github.com/aldor007/mort/commit/39cb9d5f8ffc40c5097ce6b4387a33f3ec439f83))

## [0.16.3](https://github.com/aldor007/mort/compare/v0.16.2...v0.16.3) (2022-03-15)


### Bug Fixes

* trigger release - test deployment ([ffda6a3](https://github.com/aldor007/mort/commit/ffda6a30d3c97d68b79cbc5ea634c3c996ad3942))

## [0.16.2](https://github.com/aldor007/mort/compare/v0.16.1...v0.16.2) (2022-03-15)


### Bug Fixes

* trigger release - test deployment ([5a7dde0](https://github.com/aldor007/mort/commit/5a7dde06eedd182fd68fbdf7dd6c06db78a167d6))

## [0.16.1](https://github.com/aldor007/mort/compare/v0.16.0...v0.16.1) (2022-03-15)


### Bug Fixes

* trigger release - test deployment ([2f9f05b](https://github.com/aldor007/mort/commit/2f9f05b25d5194d1245b9daf432a10ce4ece346b))

# [0.16.0](https://github.com/aldor007/mort/compare/v0.15.7...v0.16.0) (2022-03-15)


### Features

* refactor tengo script - use object.FileObject for image  manipulation ([4d1d470](https://github.com/aldor007/mort/commit/4d1d470d885da21abe2dcb1f9477f9a07f73dc74))

## [0.15.7](https://github.com/aldor007/mort/compare/v0.15.6...v0.15.7) (2022-03-15)


### Bug Fixes

* fix docker build tag v6 ([57577a1](https://github.com/aldor007/mort/commit/57577a15052c131dd65bbc85d047ed5ad2ffdf08))

## [0.15.6](https://github.com/aldor007/mort/compare/v0.15.5...v0.15.6) (2022-03-15)


### Bug Fixes

* fix docker build tag v6 ([30c1995](https://github.com/aldor007/mort/commit/30c1995e61860699f558f610ffd49f38cc229b6b))

## [0.15.5](https://github.com/aldor007/mort/compare/v0.15.4...v0.15.5) (2022-03-15)


### Bug Fixes

* fix docker build tag v6 ([cb7a527](https://github.com/aldor007/mort/commit/cb7a527f421bc46cdb0bc5aef48155fcdcdd86ef))

## [0.15.4](https://github.com/aldor007/mort/compare/v0.15.3...v0.15.4) (2022-03-15)


### Bug Fixes

* fix docker build tag v5 ([b59f335](https://github.com/aldor007/mort/commit/b59f335a1885d8f002bd8a8bfe34704d550e3328))

## [0.15.3](https://github.com/aldor007/mort/compare/v0.15.2...v0.15.3) (2022-03-15)


### Bug Fixes

* fix docker build tag v4 ([90472a8](https://github.com/aldor007/mort/commit/90472a86f6800b4e8843f55838fa8b812a9c724a))

## [0.15.2](https://github.com/aldor007/mort/compare/v0.15.1...v0.15.2) (2022-03-15)


### Bug Fixes

* fix docker build tag v3 ([4a4c17f](https://github.com/aldor007/mort/commit/4a4c17f21be9e309ed89cd1e7b7831022a28e1de))

## [0.15.1](https://github.com/aldor007/mort/compare/v0.15.0...v0.15.1) (2022-03-15)


### Bug Fixes

* fix docker build tag v2 ([41d67cd](https://github.com/aldor007/mort/commit/41d67cd2f50b19373c6f854b154b6f9d68ba5e3e))

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
