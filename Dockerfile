FROM golang:1.23.3-alpine AS builder
RUN apk update && apk add --no-cache 'git=~2'

ENV GO111MODULE=on
WORKDIR $GOPATH/src/packages/goginapp/
COPY . .

RUN go get ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/main ./cmd/main.go

FROM alpine:3

WORKDIR /

COPY --from=builder /go/main /go/main

ENV PORT=8080
ENV GIN_MODE=release

WORKDIR /go
COPY ./migrations ./migrations

ENTRYPOINT ["/go/main"]
