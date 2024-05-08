ORG    := yykhomenko
NAME   := sis
REPO   := ${ORG}/${NAME}
TAG    := ${REPO}:$(shell date '+%y%m%d')-$(shell git log -1 --pretty=format:'%h')
LATEST := ${REPO}:latest

include .env
export

update: ## Update dependencies
	go get -u ./...
	go mod tidy

lint: ## Run linters
	golangci-lint run --no-config --issues-exit-code=0 --timeout=10m \
    --disable-all --enable=deadcode  --enable=gocyclo --enable=revive --enable=varcheck \
    --enable=structcheck --enable=maligned --enable=errcheck --enable=dupl --enable=ineffassign \
    --enable=interfacer --enable=unconvert --enable=goconst --enable=gosec --enable=megacheck

test:	## Run tests
	go test -race -timeout 30s ./...

bench: ## Run benchmarks
	go test ./... -bench=. -benchmem

build: ## Build version
	go build ./cmd/${NAME}

run: ## Build and start version
	go run ./cmd/${NAME}

start: ## Start version
	./${NAME}

clean: ## Clean project
	rm -f ${NAME}
	find . -name '.DS_Store' -type f -delete

image: ## Build image
	docker build -t ${TAG} -t ${LATEST} .

pull: ## Pull image
	docker pull ${LATEST}

push: ## Push image
	docker push ${VERSION} && \
	docker push ${LATEST}

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
  {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
