run:
	go run ./cmd/order-service

build:
	go build -o bin/order ./cmd/order-service

docker:
	docker build -t order-service:local .
