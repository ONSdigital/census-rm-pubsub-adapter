up:
	docker-compose -f dependencies-compose.yml -f pubsub-adapter-compose.yml up -d;
	./setup_dependencies.sh

up-dependencies:
	docker-compose -f dependencies-compose.yml up -d;
	./setup_dependencies.sh
	
down:
	docker-compose -f dependencies-compose.yml -f pubsub-adapter-compose.yml down

docker:
	docker build -t eu.gcr.io/census-rm-ci/census-rm-pubsub-adapter .

build:
	go build -race .

format:
	go fmt ./...

format-check:
	./format_check.sh

logs:
	docker-compose -f dependencies-compose.yml -f pubsub-adapter-compose.yml logs --follow

unit-test:
	go test -race ./processor/./...

int-test: down
	docker-compose -f dependencies-compose.yml up -d;
	./setup_dependencies.sh
	PUBSUB_EMULATOR_HOST=localhost:8539 go test *.go

test: unit-test int-test

build-test: format build test
