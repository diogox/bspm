.PHONY: generate install

install:
	go install -ldflags="-s -w -X 'main.Version=local'" ./cmd/bspm

generate: clean-mock
	rm -rf vendor
	go generate ./...
	go mod vendor

clean-mock:
	find . -name '*_mock.go' -delete
