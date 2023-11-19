FROM golang:1.21.4-buster as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -mod=readonly -v -o server cmd/server/main.go

FROM gcr.io/distroless/static

COPY --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]
