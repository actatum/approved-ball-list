install-tools:
	echo Installing Tools
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

generate:
	go generate ./...

lint:
	golangci-lint run

test: 
	go test -v -race -cover -failfast ./...

build-image:
	docker build \
		-f ./Dockerfile \
		-t ${GOOGLE_COMPUTE_REGION}-docker.pkg.dev/${GOOGLE_PROJECT_ID}/abl/abl:latest \
		-t ${GOOGLE_COMPUTE_REGION}-docker.pkg.dev/${GOOGLE_PROJECT_ID}/abl/abl:${CIRCLE_SHA1} .

push-image:
	docker push ${GOOGLE_COMPUTE_REGION}-docker.pkg.dev/${GOOGLE_PROJECT_ID}/abl/abl --all-tags

start-crdb:
	./start-cockroach.sh

stop-crdb:
	./stop-cockroach.sh