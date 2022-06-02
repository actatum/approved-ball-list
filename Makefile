lint:
	golangci-lint run

test: 
	go test -v -race -cover -failfast ./...

start-crdb:
	./start-cockroach.sh

stop-crdb:
	./stop-cockroach.sh


qv-RBJVs1-qSVf_5ScLyow
postgresql://abl:qv-RBJVs1-qSVf_5ScLyow@free-tier.gcp-us-central1.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full&options=--cluster%3Dapproved-ball-list-5954