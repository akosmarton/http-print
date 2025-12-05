FROM golang:1.25.4 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v

FROM alpine

COPY --from=builder /app/connector /app/connector

ENV API_URL= API_KEY= PRINTER_URL=

ENTRYPOINT [ "/app/connector" ]
