FROM golang:1.23.3 AS builder

WORKDIR /src

COPY src/go.mod src/go.sum .
RUN go mod download

COPY src/. .

RUN GOOS=linux GOARCH=amd64 go build -o idler .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /src/idler .

EXPOSE 8080
