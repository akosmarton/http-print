FROM golang:1.25.4 as builder

WORKDIR /app
COPY . .
RUN --mount=type=cache,target=$HOME/.cache/go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v

FROM alpine

COPY --from=builder /app/http-print-server /app/http-print-server

ENV PRINTERS= APIKEY= PORT=8080

EXPOSE 8080

ENTRYPOINT [ "/app/http-print-server" ]
