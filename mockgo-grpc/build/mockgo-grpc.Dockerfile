FROM alpine:3
WORKDIR /app
RUN apk update && apk upgrade
RUN apk add --no-cache ca-certificates
COPY bin/mockgo-grpc-linux-amd64 ./mockgo-grpc
CMD ["/app/mockgo-grpc"]