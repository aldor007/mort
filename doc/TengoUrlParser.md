# Tengo URL parser

With mort you can use your own url parser for extracting operation for images

## Configuration

To enable tengo decoder you need to have configuration like below

```yaml
# config.yaml
buckets:
    tengo:
        keys:
          - accessKey: "acc"
            secretAccessKey: "sec"
        transform:
            kind: "tengo" # enable tengo decoder/parser
            tengoPath: 'parse.tengo' # path to tengo script
        storages:
            basic:
                kind: "http"
                url: "https://i.imgur.com/<item>"
                headers:
                  "x--key": "sec"
            transform:
                kind: "local-meta"
                rootPath: "/tmp/mort/"
                pathPrefix: "transforms"
```

Tengo script

```go
fmt := import("fmt")
text := import("text")

parse := func(reqUrl, bucketConfigF, obj) {
     // split by "." to remove object extension
    elements := text.split_n(reqUrl.path, ".", 2)
    ext := elements[1]
    if len(elements) == 1 {
        return ""
    }
    // split by "," to find resize parameters
    elements = text.split(elements[0], ",")

    // url has no transform
    if len(elements) == 1 {
        return ""
    }

    // apply parameters
    width := 0
    height := 0
    parent := elements[0] +"." +  ext
    trans := elements[1:]
    for tran in trans {
        if tran[0] == 'w' {
            width = tran[1:]
        }

        if tran[0] == 'h' {
            height = tran[1:]
        }
    }

    obj.transforms.resize(int(width), int(height), false, false, false)
    return parent
}

parent := parse(url, bucketConfig, obj)
err := undefined
```

Above script will work for URL http://localhost:8084/tengo/udXmD2T,w100,h100.jpeg


Mort is injecting variables inside of tengo script
* `url` golang net.URL struct
* `bucketConfig` - mort bucket configuration
* `obj` - mort object.FileObject
** `obj.transforms` - mort transform.Transforms object on which you can execute image manipulations

Output variables
* `parent` - path for parent object
* `err` - error if occurred
