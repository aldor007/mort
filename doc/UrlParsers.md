# Writing you own url decoder

Using mort you can write you own request url parser. This section will guide you how to do so and use it.

## Implementation

Parser should be function with look like
```go
type ParseFnc func(url *url.URL, bucketConfig config.Bucket, obj *FileObject) (parent string, err error)

```
Function parameters:
* url is a request URI
* bucketConfig - is configuration for current bucket
* obj - is result object. In which you should store operation to perform

It should return
* parent - string path for original image
* error - if any error occurred

To register parser you should call 
```go
object.RegisterParser(kind string, fn ParserFNc)
```
Parser have to be registered before loading configuration. 
It will be called on transform object and original object.

# Example custom parser

This parser will parse url like
https://mort/bucket/parent,w100,h100.png


```go
object.RegisterParser("custom", func (reqUrl *url.URL, bucketConfig config.Bucket, obj *object.FileObject)  (string, error) {
    // split by "." to remove object extension
    elements := strings.SplitN(reqUrl.Path, ".", 2)
    if len(elements) == 1 {
        return "", nil
    }
    // split by "," to find resize parameters
    elements = strings.Split(elements[0], ",")

    // url has no transform
    if len(elements) == 1 {
        return "", nil
    }

    // apply parameters
    var width, height int64
    parent := elements[0] +  path.Ext(reqUrl.Path)
    trans := elements[1:]
    for _, tran := range trans {
        if tran[0] == 'w' {
            width, _ = strconv.ParseInt(tran[1:], 10, 32)
        }

        if tran[0] == 'h' {
            height, _ = strconv.ParseInt(tran[1:], 10, 32)
        }
    }

    obj.Transforms.Resize(int(width), int(height), false)
    return parent, nil
})
```

In bucket configuration:
```yaml
buckets:
    bucket:
      transform:
        kind: "custom"
```

