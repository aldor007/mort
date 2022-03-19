# Mort
[![Tests](https://github.com/aldor007/mort/actions/workflows/ci.yaml/badge.svg)](https://github.com/aldor007/mort/actions/workflows/ci.yaml) [![Codecov](https://img.shields.io/codecov/c/github/aldor007/mort.svg)](https://codecov.io/gh/aldor007/mort) [![Go Report Card](https://goreportcard.com/badge/github.com/aldor007/mort)](https://goreportcard.com/report/github.com/aldor007/mort)  [![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/aldor007/mort) [![Releases](https://img.shields.io/github/release/aldor007/mort/all.svg?style=flat-square)](https://github.com/aldor007/mort/releases)  [![LICENSE](https://img.shields.io/github/license/aldor007/mort.svg?style=flat-square)](https://github.com/aldor007/mort/blob/master/LICENSE.md)

<img src="https://mort.mkaciuba.com/demo/medium/gopher.png" width="500px"/>


An S3-compatible image processing server written in Go. Still in active development.

# Features

* HTTP server
* Resize, Rotate, SmartCrop
* Convert (JPEG, PNG , BMP, Webp)
* Multiple storage backends (disk, S3, http)
* Fully modular
* S3 API for listing and uploading files
* Requests collapsing
* Build in rate limiter
* HTTP Range and Conditional requests
* Compression (gzip, brotli)
* Env variables driven configuration
* Write you own transform parser using [tengo](https://github.com/d5/tengo)
* cloudinary image transformation and upload

And more see [changelog](CHANGELOG.md) for more info

# Demo
-------
[Original image](https://mort.mkaciuba.com/demo/img.jpg)

Click on result image to see URL. More examples can be found in [Image Operations list](doc/Image-Operations.md)
<table>
    <thead>
    <tr>
        <th>Description</th>
        <th>Result (to see result click on image)</th>
     </tr>
    </thead>
    <tbody>
        <tr>
            <td>
                <p>preset: small</p>
                <p>(preserve aspect ratio)
                 width: 75 </p>
            </td>
            <td>
               <a href="https://mort.mkaciuba.com/demo/small/img.jpg" target="_blank">
                <img src="https://mort.mkaciuba.com/demo/small/img.jpg" width="75px" />
                </a>
            </td>
        </tr>
        <tr>
            <td>
                <p>preset: blur</p>
                <ul>
                <li><p>resize image (preserve aspect ratio)
                 width: 700</p></li>
                 <li><p>blur image with sigma 5.0</p></li>
                 </ul>
            </td>
            <td>
                  <a href="https://mort.mkaciuba.com/demo/blur/img.jpg" target="_blank">
                <img src="https://mort.mkaciuba.com/demo/blur/img.jpg" width="700px" />
                </a>
            </td>
        </tr>
        <tr>
            <td>
                <p>preset: webp</p>
                <ul>
                <li>                <p>resize image (preserve aspect ratio)
                 width: 1000</p></li>
                 <li><p>and change format to webp</p></li>
                 </ul>
            </td>
            <td>
                <a href="https://mort.mkaciuba.com/demo/webp/img.jpg" target="_blank">
               <img src="https://mort.mkaciuba.com/demo/webp/img.jpg" width="1000px" />
                </a>
            </td>
        </tr>
        <tr>
            <td>
                <p>preset: watermark</p>
                <ul>
                <li>                <p>resize image (preserve aspect ratio)
                 width: 1300</p></li>
                 <li><p>and add watermark</p></li>
                 </ul>
            </td>
            <td>
                <a href="https://mort.mkaciuba.com/demo/watermark/img.jpg" target="_blank">
                  <img src="https://mort.mkaciuba.com/demo/watermark/img.jpg" width="1300px" />
                </a>
            </td>
        </tr>
     </tbody>
</table>

# Usage

Mort can be used directly from the Internet and behind any proxy.


## Install

```bash
go get github.com/aldor007/mort/cmd/
```


## Command line help
```bash
$ ./mort
Usage of  mort
  -config string
    	Path to configuration (default "/etc/mort/mort.yml")
```

## Configuration
Example configuration used for providing demo images:

```yaml
headers: #  add or overwrite all response headers of given status. This field is optional
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
                blur:
                    quality: 80
                    filters:
                        thumbnail:
                            width: 700
                        blur:
                          sigma: 5.0
                webp:
                    quality: 100
                    format: webp
                    filters:
                        thumbnail:
                            width: 1000
                watermark:
                    quality: 100
                    filters:
                        thumbnail:
                            width: 1300
                        watermark:
                            image: "https://i.imgur.com/uomkVIL.png"
                            position: "top-left"
                            opacity: 0.5
                smartcrop:
                    quality: 80
                    filters:
                      crop:
                        width: 200
                        height: 200
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
                 pathPrefix: "transform"

```
List of all image operations can be found in [Image-Operations.md](doc/Image-Operations.md)

More details about configuration can be found in [Configuration.md](doc/Configuration.md)


## Write you own URL parser

- Using golang (custom mort version needed) [UrlParsers.md](doc/UrlParsers.md)
- Using tengo [TengoUrlParser.md](doc/TengoUrlParser.md)

## Docker
See [Dockerfile](Dockerfile) for image details.

Pull docker image

```bash
docker pull ghcr.io/aldor007/mort

```

### Create you custom docker deployment

Create Dockerfile or use Dockerfile.service
```
# Dockerfile.service 
FROM ghcr.io/aldor007/mort:latest

# add your config
ADD config.yml /etc/mort/mort.yml
```

Build container
```bash
docker build -f Dockerfile.service -t myusername/mort
```

Run docker
```
docker run -p 8080:8080 myusername/mort
```

Full example you can find [here](example/)

# Development
1. Make sure you have a Go language compiler >= 1.9 (required) and git installed.
2. Install libvips like described on [bimg page](https://github.com/h2non/bimg)
3. Ensure your GOPATH is properly set.
4. Download it
```bash
git clone  https://github.com/aldor007/mort.git $GOPATH/src/github.com/aldor007/mort
cd $GOPATH/src/github.com/aldor007/mort
```
5. Install dependencies:
```bash
go mod vendor
```
Run unit tests:
```bash
make unit
```
Run integration tests:
```bash
make integrations
```

If you encounter following problem `go build gopkg.in/h2non/bimg.v1: invalid flag in pkg-config --cflags: -Xpreprocessor`
while running tests in your IDE add this to your environment:
```
export CGO_CFLAGS_ALLOW="-Xpreprocessor"
```

Run tests using docker image
```bash
make docker-tests
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Inspirations

* [picfit](https://github.com/thoas/picfit)
* [imaginary](https://github.com/h2non/imaginary)

