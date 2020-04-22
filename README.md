# census-rm-pubsub-adapter
[![Build Status](https://travis-ci.com/ONSdigital/census-rm-pubsub-adapter.svg?branch=master)](https://travis-ci.com/ONSdigital/census-rm-pubsub-adapter)

An adapter service to translate inbound PubSub messages into the standard format of RM JSON events and republish them on to our events exchange.

## Prerequisites 
Requires golang >= 1.13 installed

## Running the tests
Run 
```shell
make test
```
This will run the units tests then spin up the dependencies with docker-compose and run the service integration tests.

## Debugging the tests
To run the integration tests in an IDE
 1. Run `make up` to start up the dependencies with docker-compose.
 1. Set the environment variable `PUBSUB_EMULATOR_HOST=localhost:8539` in your IDE run configuration
 1. Run the test in debug mode

## Formatting
Run `make format` to automatically format the project using `gofmt`
