FROM golang:1.16-alpine

RUN apk update && apk add git

RUN go get github.com/gomodule/redigo/redis

ADD . /usr/local/go/src/visit-counter
WORKDIR /usr/local/go/src/visit-counter
RUN go install visit-counter

ENV REDISHOST redis
ENV REDISPORT 6379

USER 1000

ENTRYPOINT /usr/local/go/bin/visit-counter

EXPOSE 8080
