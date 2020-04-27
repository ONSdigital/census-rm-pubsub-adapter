up:
	docker-compose up -d;
	./setup_dependencies.sh

up-dependencies:
	docker-compose up -d rabbitmq pubsub-emulator;
	./setup_dependencies.sh
	
down:
	docker-compose down

docker:
	docker build -t eu.gcr.io/census-rm-ci/census-rm-pubsub-adapter .

build:
	go build -race .

format:
	go fmt ./...

format-check:
	./format_check.sh

logs:
	docker-compose logs --follow

unit-test:
	go test -race ./... -tags=unitTest

int-test: down up-dependencies
	PUBSUB_EMULATOR_HOST=localhost:8539 go test .
	docker-compose down;

test: unit-test int-test

build-test: format build test docker
