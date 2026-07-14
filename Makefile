.PHONY: generate test lint test-race test-golden-update clean

# Generate code for examples.
generate:
	go generate ./...

# Run tests.
test:
	go test ./... -timeout 120s

# Run tests with race detector.
test-race:
	go test -race ./... -timeout 120s

# Lint using golangci-lint.
lint:
	golangci-lint run ./...

# Tidy go modules.
tidy:
	go mod tidy

# Clean build artifacts.
clean:
	rm -rf _examples/kitchensink/ent
