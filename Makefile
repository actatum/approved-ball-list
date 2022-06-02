lint:
	golangci-lint run

test: 
	go test -v -race -cover -failfast ./...

start-crdb:
	./start-cockroach.sh

stop-crdb:
	./stop-cockroach.sh