FROM golang

WORKDIR /go/src/http-print-server
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

WORKDIR /data

VOLUME /data

ENV APIKEY ""
ENV DBPATH jobs.db
ENV PORT 8080

EXPOSE 8080

CMD ["/go/bin/http-print-server"]
