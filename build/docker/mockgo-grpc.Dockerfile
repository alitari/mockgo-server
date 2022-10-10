FROM golang:1.19.1 as builder
WORKDIR /app

# protobuf
RUN apt update -y && apt install -y protobuf-compiler 
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest 

# mockgo core
COPY ./mockgo ./mockgo

# kvstore
COPY ./grpc-kvstore ./grpc-kvstore
COPY scripts/go-build-grpc-kvstore.sh .
RUN ./go-build-grpc-kvstore.sh
# matchstore
COPY ./grpc-matchstore ./grpc-matchstore
COPY scripts/go-build-grpc-matchstore.sh .
RUN ./go-build-grpc-matchstore.sh

# mockgo main
COPY ./mockgo-grpc ./mockgo-grpc
RUN echo "go 1.19\n\nuse (\n    ./mockgo\n    ./grpc-kvstore\n    ./grpc-matchstore\n    ./mockgo-grpc\n)" > ./go.work
RUN go mod download
COPY scripts/go-build-mockgo-grpc.sh .
RUN ./go-build-mockgo-grpc.sh linux amd64

FROM alpine:3
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/mockgo-grpc-linux-amd64 ./mockgo-grpc
CMD ["/app/mockgo-grpc"]