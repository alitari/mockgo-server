FROM golang:1.19.1 as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY scripts/go-build.sh .
COPY test/main ./test
RUN ./go-build.sh

FROM alpine:3
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/mockgo-server ./mockgo-server
COPY --from=builder /app/test .
CMD ["/app/mockgo-server"]