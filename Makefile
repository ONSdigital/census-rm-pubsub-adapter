up:
	docker-compose -f dev.yml up -d;
	./setup_dependencies.sh
	
down:
	docker-compose -f dev.yml down

docker:
	docker build -race -t eu.gcr.io/census-rm-ci/census-rm-pubsub-adapter .

build:
	go build .

format:
	go fmt ./...

format-check:
	./format_check.sh

logs:
	docker-compose -f dev.yml logs --follow

unit-test:
	go test -race ./processor/./...

int-test:
	docker-compose -f dev.yml up -d;
	./setup_dependencies.sh
	PUBSUB_EMULATOR_HOST=localhost:8539 go test *.go

test: unit-test int-test
