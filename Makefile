lint:
	goimports-reviser -rm-unused -set-alias -format ./...

build:
	go build -o story-staking-api ./cmd/main.go

run:
	go run ./cmd/main.go

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o story-staking-api ./cmd/main.go

test:
	go clean -testcache && go test ./... -covermode=atomic

gen-sqlc:
	pushd schema && sqlc generate && popd

local-postgres:
	docker pull postgres:latest
	docker run --restart=always \
		--name story-staking-api-postgres \
		-p 5432:5432 \
		-e POSTGRES_USER=my-user \
		-e POSTGRES_PASSWORD=my-secret \
		-e POSTGRES_DB=postgres \
		-d postgres:latest

local-redis:
	docker pull redis:latest
	docker run -itd --name story-staking-api-redis -p 6379:6379 redis
