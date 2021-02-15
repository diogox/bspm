.PHONY: generate

generate: clean-mock
	rm -rf vendor
	go generate ./...
	go mod vendor

clean-mock:
	find . -name '*_mock.go' -delete
