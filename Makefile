.PHONY: test test-domain mocks run

test:
	go test ./...

test-domain:
	go test ./internal/domain -v

mocks:
	go run ./cmd/mock-upstreams

run:
	go run ./cmd/playback-api
