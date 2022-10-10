FROM golang:1.19.1 as builder
WORKDIR /app
COPY ./mockgo ./mockgo
COPY ./mockgo-standalone ./mockgo-standalone
RUN echo "go 1.19\n\nuse (\n	./mockgo\n	./mockgo-standalone\n)" > ./go.work
RUN go mod download
COPY scripts/go-build-mockgo-standalone.sh .
RUN ./go-build-mockgo-standalone.sh linux amd64

FROM alpine:3
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/mockgo-standalone-linux-amd64 ./mockgo-standalone
CMD ["/app/mockgo-standalone"]