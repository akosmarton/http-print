FROM golang:1.17 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v

FROM alpine

COPY --from=builder /app/server /app/server

VOLUME /data
ENV APIKEY= DBPATH=/data/jobs.db PORT=8080

EXPOSE 8080

ENTRYPOINT [ "/app/server" ]
