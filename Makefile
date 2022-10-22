build: ## Build version
	go build -v ./cmd/sis

test: ## Run all tests
	go test -race -timeout 10s ./...

bench: ## Run all benchmarks
	go test ./... -bench=. -benchmem

run: ## Run version
	go run ./cmd/sis

install: ## Install version
	make build
	make test
	go install -v ./cmd/sis

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'
