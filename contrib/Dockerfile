FROM golang:1.16-alpine as builder

COPY . /app

WORKDIR /app

ARG VERSION "0.0.0-dev"

RUN CGO_ENABLED=0 go build -o myback -ldflags "-s -w -X github.com/axllent/myback/cmd.Version=${VERSION} -X github.com/axllent/myback/cmd.IsDocker=true" . && \
apk add --no-cache upx && upx -9 myback


FROM alpine

RUN apk add --no-cache tzdata mariadb-client bash

COPY --from=builder /app/myback /usr/local/bin/

ENTRYPOINT ["myback"]