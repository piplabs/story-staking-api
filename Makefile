lint:
	@golangci-lint run --allow-parallel-runners

lint-fix:
	@golangci-lint run --fix --allow-parallel-runners

run:
	go run ./cmd/main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o story-indexer ./cmd/main.go

test:
	go clean -testcache && go test ./... -covermode=atomic

gen-sqlc:
	pushd schema && sqlc generate && popd

local-postgres:
	docker pull postgres:latest
	docker run --restart=always \
		--name story-indexer-postgres \
		-p 5432:5432 \
		-e POSTGRES_USER=my-user \
		-e POSTGRES_PASSWORD=my-secret \
		-e POSTGRES_DB=postgres \
		-d postgres:latest

local-redis:
	docker pull redis:latest
	docker run -itd --name story-indexer-redis -p 6379:6379 redis
