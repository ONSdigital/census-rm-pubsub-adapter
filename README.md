# census-rm-pubsub-adapter
[![Build Status](https://travis-ci.com/ONSdigital/census-rm-pubsub-adapter.svg?branch=master)](https://travis-ci.com/ONSdigital/census-rm-pubsub-adapter)

An adapter service to translate inbound PubSub messages into the standard format of RM JSON events and republish them on to our events exchange.

## Prerequisites 
Requires golang >= 1.13 installed

## Configuration

The required environment configuration variables are:
```sh
RABBIT_HOST
RABBIT_PORT
RABBIT_USERNAME
RABBIT_PASSWORD
EQ_RECEIPT_PROJECT
EQ_RECEIPT_SUBSCRIPTION
```

Development config to run locally using the docker compose dependencies is provided in [development.env](./development.env)

## Running the tests
Run 
```sh
make build-test
```
This will run the formatter, build and units tests then spin up the dependencies with docker-compose and run the service integration tests.

## Debugging the tests
To run the integration tests in an IDE
 1. Run `make up-dependencies` to start up the dependencies with docker-compose.
 1. Set the environment variable `PUBSUB_EMULATOR_HOST=localhost:8539` in your IDE run configuration
 1. Run the test in debug mode

## Formatting
Run `make format` to automatically format the project using `gofmt`

## Build the docker image
With 
```sh
make docker
```    

## Run in docker-compose
### Start the service and dependencies
Run `make up` to start the pubsub-adapter and dependencies through docker-compose

You can then run `make logs` to tail the logs

### Post in a test message
You can send a test message onto the pubsub emulator with the tools script
```sh
PUBSUB_EMULATOR_HOST=localhost:8539 go run tools/publish_message.go
```
You should see the pubsub adapter log that it has processed the message and see the rabbit messages it produced in the rabbit management UI at http://localhost:17672 (login: guest, guest).
