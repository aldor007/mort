# Mort [![Build Status](https://travis-ci.org/aldor007/mort.png)](https://travis-ci.org/aldor007/mort) [![Docker](https://img.shields.io/badge/docker-aldor007/mort-blue.svg)](https://hub.docker.com/r/aldor007/mort/) [![Docker Registry](https://img.shields.io/docker/pulls/aldor007/mort.svg)](https://hub.docker.com/r/aldor007/mort/) [![Go Report Card](http://goreportcard.com/badge/aldor007/mort)](http://goreportcard.com/report/aldor007/mort) 

S3 compatible image processing server written in Go.

# Features

* HTTP server
* Resize 
* Rotate
* SmartCrop
* Convert (JPEG, , PNG , BMP, TIFF, ...)
* Multiple storage backends (disk, S3, http)
* Fully modular
* S3 API for listing and uploading files 

# Demo
-------
[Original image](https://mort.mkaciuba.com/demo/cat.jpg)
<table>
    <thead>
    <tr>
        <th>Description</th>
        <th>Result</th> 
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
                <a href="https://mort.mkaciuba.com/demo/small/cat.jpg" target="_blank">
                <img src="https://mort.mkaciuba.com/demo/small/cat.jpg">
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
                <a href="https://mort.mkaciuba.com/demo/blur/cat.jpg" target="_blank">
                <img src="https://mort.mkaciuba.com/demo/blur/cat.jpg">
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
                <a href="https://mort.mkaciuba.com/demo/webp/cat.jpg" target="_blank">
                <img src="https://mort.mkaciuba.com/demo/webp/cat.jpg">
                </a>
            </td>
        </tr>
     </tbody>
</table>   
    
# Usage

Mort can be used direct from Internet and behind any proxy. 

## Command line help
```bash
$ ./mort
Usage of  mort
  -config string
    	Path to configuration (default "configuration/config.yml")
  -listen string
    	Listen addr (default ":8080")
```

## Configuration
Example configuration used for providing demo images:

```yaml
headers: # overwritten all response headers of given status. This field is optional
  - statusCodes: [200]
    values:
      "cache-control": "max-age=84000, public"

buckets: # list of available buckets 
    demo:    # bucket name 
        keys: # list of S3 keys (optiona
          - accessKey: "access"
            secretAccessKey: "random"
        transform: # config for transforms
            path: "\\/(?P<presetName>[a-z0-9_]+)\\/(?P<parent>[a-z0-9-\\.]+)" # regexp for transform path 
            kind: "presets" #  type of transform for now only "presets" is available 
            presets: # list of presets
                small:
                    quality: 75
                    filters:
                        thumbnail: {size: [150]}
                blur:
                    quality: 80
                    filters:
                        thumbnail: {size: [700]}
                        blur:
                          sigma: 5.0
                webp:
                    quality: 100
                    format: webp
                    filters:
                        thumbnail: {size: [1000]}
                watermark:
                    quality: 100
                    filters:
                        thumbnail: {size: [1300]}
                        watermark:
                            image: "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e9/Imgur_logo.svg/150px-Imgur_logo.svg.png"
                            position: "center-center"
                            opacity: 0.5
        storages:
             basic: # retrieve originals from s3
                 kind: "s3"
                 accessKey: "acc"
                 secretAccessKey: "sec"
                 region: ""
                 endpoint: "http://localhost:8080"
             transform: # end stroe it on disk
                 kind: "local-meta"
                 rootPath: "/var/www/domain/"
                 pathPrefix: "transform"
        
```


## Debian and Ubuntu

We will provide Debian package when we will be completely stable ;)

## Docker
Pull docker image

```bash
docker pull aldor007/mort

```

Create Dockerfile
```
FROM aldor007/mort:latest
ADD config.yml /go/configuration/config.yml # add you configu
```

Run docker 


# Development
1. Make sure you have a Go language compiler >= 1.9 (required) and git installed.
2. Install libvips like described on [bimg page](https://github.com/h2non/bimg)
3. Ensure your GOPATH is properly set.
4. Download it
```bash
go get -d github.com/aldor007/mort
cd $GOPATH/src/github.com/aldor007/mort
```
5, Install dependencies:
```bash
dep ensure
```
Run end to end tests:
```bash
make unit
```
Run integration tests:
```bash
make integrations
```

## Built With

* [dep](https://github.com/golang/dep) - Dependency Management
* [bimg](https://github.com/h2non/bimg) -  Image processing powered by libvips C library

## Contributing

Please read [CONTRIBUTING.md](https://github.com/aldor007/mort/CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/aldor007/mort/tags). 

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Inspirations

* [picfit](https://github.com/thoas/picfit) 
* [imaginary](https://github.com/h2non/imaginary)

