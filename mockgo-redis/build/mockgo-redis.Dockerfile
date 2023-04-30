FROM alpine:3
WORKDIR /app
RUN apk update && apk upgrade
RUN apk add --no-cache ca-certificates
COPY ./cmd/bin/mockgo-redis-linux-amd64 ./mockgo-redis
CMD ["/app/mockgo-redis"]