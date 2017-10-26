# Mort [![Build Status](https://travis-ci.org/aldor007/mort.png)](https://travis-ci.org/aldor007/mort) [![Docker](https://img.shields.io/badge/docker-aldor007/mort-blue.svg)](https://hub.docker.com/r/aldor007/mort/) [![Docker Registry](https://img.shields.io/docker/pulls/aldor007/mort.svg)](https://hub.docker.com/r/aldor007/mort/) [![Go Report Card](http://goreportcard.com/badge/aldor007/mort)](http://goreportcard.com/report/aldor007/mort) 


S3 compatible image processing server written in Go.

# Usage

# Configuration


## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

What things you need to install the software and how to install them

```
Give examples
```

### Installing

A step by step series of examples that tell you have to get a development env running

Say what the step will be

```
Give the example
```

And repeat

```
until finished
```

End with an example of getting some data out of the system or using it for a little demo

## Running the tests

Explain how to run the automated tests for this system

### Break down into end to end tests

Explain what these tests test and why

```
Give an example
```

### And coding style tests

Explain what these tests test and why

```
Give an example
```

## Deployment

```bash
git clone https://github.com/aldor007/mort.git && cd mort 
```
Install dependencies:
```bash
dep ensure
```
Run tests:
```bash
go test ./...
```
Run integration tests:
```bash
./run-int.sh
```

## Built With

* [echo](https://github.com/labstack/echo) - The web framework used
* [dep](https://github.com/golang/dep) - Dependency Management
* [bimg](https://github.com/h2non/bimg) -  Image processing powered by libvips C library

## Contributing

Please read [CONTRIBUTING.md](https://github.com/aldor007/mort/CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/aldor007/mort/tags). 

## Authors

* **Marcin Kaciuba** - *Initial work* - [Aldor007](https://github.com/aldor007)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Inspirations

* [picfit](https://github.com/thoas/picfit) 
* [imaginary](https://github.com/h2non/imaginary)

