lint:
	golangci-lint run

test: 
	go test -v -race -cover -failfast ./...
