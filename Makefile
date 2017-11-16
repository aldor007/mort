
install:
	dep ensure

unit:
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

e2e:
	npm install
	./scripts/run-tests.sh

format:
	@(go fmt ./...)
	@(go vet ./...)

test: unit e2e
