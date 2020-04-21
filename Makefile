up:
	docker-compose -f dev.yml up -d;
	./setup_pubsub.sh
	
down:
	docker-compose -f dev.yml down

logs:
	docker-compose -f dev.yml logs --follow

unit-test:
	go test -race ./processor/./...

int-test:
	docker-compose -f dev.yml up -d;
	./setup_pubsub.sh
	PUBSUB_EMULATOR_HOST=localhost:8539 go test *.go

test: unit-test int-test
