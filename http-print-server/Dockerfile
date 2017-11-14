FROM golang:alpine

RUN apk update && apk add git
RUN go get github.com/akosmarton/http-print/http-print-server

WORKDIR /data

VOLUME /data

EXPOSE 8080

ENTRYPOINT ["/go/bin/http-print-server"]
