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

clean: ## Clean project
	make stop
	rm -rf db/data/{*,.[\!]*}
	find . -name '.DS_Store' -type f -delete

image: ## Build image
	docker build -t ${TAG} -t ${LATEST} .

push: ## Push image
	docker push ${TAG} && \
	docker push ${LATEST}

pull: ## Pull images
	docker compose pull

build: ## Build containers
	docker compose build

start: ## Create and start containers
	docker compose up -d

stop: ## Stop and remove containers
	docker compose down

restart: ## Restart containers
	make stop build start

status: ## Print containers status
	docker compose ps

log: ## Print log
	docker compose logs -f

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
  {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
