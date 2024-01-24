.DEFAULT_GOAL := generate

generate:
	protoc --go_out=pkg --go_opt=paths=source_relative --go-grpc_out=pkg \
     	--go-grpc_opt=paths=source_relative \
     	proto/shortener.proto

test:
	go test ./... -v

build_and_run:
	docker-compose up -d --build

run:
	docker-compose up -d