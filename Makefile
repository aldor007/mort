
install:
	dep ensure

unit: format
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -race  -cover)

unit-bench:
	./scripts/unit-travis.sh

coverage:
	go test github.com/aldor007/mort/... -race -coverprofile=coverage.txt -covermode=atomic

integrations:
	npm install
	./scripts/run-tests.sh

format:
	@(go fmt ./...)
	@(go vet ./...)

tests: unit integrations

docker-push:
	docker buildx build --platform linux/amd64,linux/arm64 -t aldor007/mort -f Dockerfile . -t aldor007/mort:latest --push; docker push aldor007/mort:latest
run-server:
	mkdir -p /tmp/mort
	go run cmd/mort/mort.go -config configuration/config.yml

run-test-server:
	scripts/prepare-for-tests.sh
	go run cmd/mort/mort.go -config ./tests-int/mort.yml

run-test-server-redis:
	scripts/prepare-for-tests.sh
	go run cmd/mort/mort.go -config ./tests-int/mort-redis.yml

clean-prof:
	find . -name ".test" -depth -exec rm {} \;
	find . -name ".cpu" -depth -exec rm {} \;
	find . -name ".mem" -depth -exec rm {} \;

release:
	docker build . -f Dockerfile.build --build-arg GITHUB_TOKEN=${GITHUB_TOKEN}

docker-tests:
	scripts/prepare-for-tests.sh
	docker-compose up --build

base-docker-push:
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/aldor007/mort-base:1.1.0-8.11.2  -f Dockerfile.base . -t ghcr.io/aldor007/mort-base:latest --push