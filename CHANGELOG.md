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
