# url
Object in which you can check url of request

# properties

* `host` - uri host
* `scheme` - uri scheme
* `path` - uri path
* `rawquery` - uri whole query in string
* `uri.query` - uri query string as map

Example usage

```go
text := import("text")
elements := text.split_n(url.path, ".", 2)
``