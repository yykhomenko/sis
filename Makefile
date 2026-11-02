ORG    := yykhomenko
NAME   := sis
REPO   := ${ORG}/${NAME}
TAG    := ${REPO}:$(shell date '+%y%m%d')-$(shell git log -1 --pretty=format:'%h')
LATEST := ${REPO}:latest

-include .env
export

update: ## Update dependencies
	go get -u ./...
	go mod tidy

lint: ## Run linters (golangci-lint v2, config in .golangci.yml)
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

test:	## Run tests
	go test -race -timeout 60s ./...

bench: ## Run benchmarks
	go test ./... -bench=. -benchmem

build_exec:
	go build -o sis ./cmd/sis

run:
	make build_exec && ./sis

clean: ## Clean project (removes local database data!)
	@read -p "This deletes all data in db/data. Continue? [y/N] " ans && [ "$$ans" = y ]
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

get_subscriber:
	curl http://localhost:9001/subscribers/380670000001

bench_sis:
	~/src/go/wrkb/wrkb -p=sis http://127.0.0.1:9001/subscribers/__RANDI64_380670000001_380670099999__

migration_up:
	@cd sql/schema && \
	goose postgres ${SIS_DB_URL} up

migration_down:
	@cd sql/schema && \
	goose postgres ${SIS_DB_URL} down

generate_database_code:
	@sqlc generate

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
  {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
