tag := $(shell git describe)

install:
	dep ensure

unit:
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

coverage:
	go test github.com/aldor007/mort/... -race -coverprofile=coverage.txt -covermode=atomic

integrations:
	npm install
	./scripts/run-tests.sh

format:
	@(go fmt ./...)
	@(go vet ./...)

tests: unit integrations

docker-build:
	docker build -t aldo007/mort -f Dockerfile .

docker-push: docker-build
	docker push aldor007/mort:latest

run-server:
	go run cmd/mort/mort.go -config configuration/config.yml
	
run-test-server:
	mkdir -p /tmp/mort-tests/remote
	mkdir -p /tmp/mort-tests/remote-query
	mkdir -p /tmp/mort-tests/local
	go run cmd/mort/mort.go -config ./tests-int/config.yml

clean-prof:
	find . -name ".test" -depth -exec rm {} \;
	find . -name ".cpu" -depth -exec rm {} \;
	find . -name ".mem" -depth -exec rm {} \;

release:
	docker build . -f Dockerfile.build --build-arg GITHUB_TOKEN=${GITHUB_TOKEN}
	docker build . -t aldor007/mort:${tag}; docker push aldor007/mort:${tag}
