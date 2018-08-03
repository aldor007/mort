* 0.12.0    
    * Feature: Add new monitoring metrics (time of image generation and count of it)
    * Feature: Do eror placeholder in background (returns it faster to user)
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
