.PHONY: diagrams build logs dummy redis

diagrams:
	./scripts/diagrams.sh

build: bin/redisTests

bin/redisTests: main.go
	go build -o bin/ .

resetter: bin/redisTests
	./bin/redisTests --resetter

consumer: bin/redisTests
	./bin/redisTests >log/worker1.log 2>&1 & \
	./bin/redisTests >log/worker2.log 2>&1 & \
	./bin/redisTests >log/worker3.log 2>&1 &

dummy:
	json-server --ro -q ./dummy-service.json

logs:
	./scripts/logs.sh

redis:
	docker run -ti --hostname redis --name redis --rm -p 6379:6379 -d redis
