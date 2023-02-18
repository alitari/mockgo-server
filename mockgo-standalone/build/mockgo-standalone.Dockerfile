FROM alpine:3
WORKDIR /app
RUN apk update && apk upgrade
RUN apk add --no-cache ca-certificates
COPY ./bin/mockgo-standalone-linux-amd64 ./mockgo-standalone
CMD ["/app/mockgo-standalone"]