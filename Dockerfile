FROM golang:latest AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o card-manager .

FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /build/card-manager .

ENTRYPOINT ["/app/card-manager"]
