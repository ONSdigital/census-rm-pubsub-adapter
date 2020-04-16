up:
	docker-compose -f dev.yml up -d;
	./setup_pubsub.sh
	
down:
	docker-compose -f dev.yml down

pull:
	docker-compose -f dev.yml pull

logs:
	docker-compose -f dev.yml logs --follow

test:
	PUBSUB_EMULATOR_HOST=localhost:8539 go test