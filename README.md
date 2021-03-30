# Distributed Client Rate Limiting

> ## NOTE: 
> 
> See [Docs](doc/RATELIMIT.md) for an explanation on how
> this works and the logic behind it.
> 
> &nbsp;

You'll need Docker, PlantUML & json-server.

## Makefile


See the [Makefile](Makefile), using it you can:

* `make diagrams` - Update generated images from PlantUML diagram definitions
* `make build` - Build the POC
* `make resetter` - Run the Resetter
* `make consumer` - Run the Consumers
* `make logs` - Display the logs
* `make dummy` - Start thw dummy json-server service
* `make redis` - Start Redis using Docker

## PlantUML

Install it with:

```shell
brew install plantuml
```

## How to install json-server

First install json-server with:

```shell
npm install -g json-server
```
