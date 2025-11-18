FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service ./cmd

FROM alpine:3.20

WORKDIR /app

ENV GIN_MODE=release

COPY --from=builder /app/service /app/service

EXPOSE 8080

CMD ["/app/service"]


