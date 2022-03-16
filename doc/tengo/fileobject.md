
# file object
This object is injected into script and available via `obj` value

## read only properties

* `uri` - net/url object which contains whole url
** `uri.host` - uri host
** `uri.scheme` - uri scheme
** `uri.path` - uri path
** `uri.rawquery` - uri whole query in string
** `uri.query` - uri query string as map
* `bucket` - string name of bucket for object
* `key` - string, storage path for object
* `transforms` - object on which you can add image manipulations. For more see [Transforms](#transform)s

## properties that can be changed

* `allowChangeKey` - bool (default: true) if storage path for object can be changed
* `checkParent` - bool (default: false) if mort should make check if parent exist before generating image
* `debug` - bool (default: false) add debug headers to response

Example usage

```go
fmt := import("fmt")
fmt.println(obj.key)
fmt.println(obj.uri)

```

# transforms

Object on which you can execute image manipulation described in [Image-Operations](doc/Image-Operations.md)

# properties

* `resize(width int, height, int, enlarge bool, preverseAspectRatio bool, fill bool)` - resize image
* `extract(top, left, width, height int)` - crop image
* `crop(width int, height int, gravity string, enlarge bool, embed bool)` - crop image
* `resizeCropAuto(width int, height int)` - crop image
* `interlace()`
* `quality(quality int)` - image quality
* `stripMetadata()` - remove metadata
* `blur(sigma float, mingAmpl float)` - blur image
* `format(format string)` - change image format
* `watermark(image string, position string, opacity float)` - add watermark to image
* `grayscale()` - image in grayscale
* `rotate(angle int)` - rotate image

