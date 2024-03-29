package kvstore

import (
	context "context"
	"encoding/json"
	"fmt"
	"github.com/alitari/mockgo-server/mockgo/util"
	"go.uber.org/zap"
	"log"
	"net"
	"time"

	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcStorage struct {
	id string
	*kvstore.InmemoryStorage
	clients []KVStoreClient
	timeout time.Duration
	logger  *zap.Logger
	UnimplementedKVStoreServer
	server *grpc.Server
}

/*
NewGrpcStorage creates a new distributed kvstore.Storage.
*/
func NewGrpcStorage(addresses []string, serverPort int, logLevel string) (kvstore.Storage, error) {
	storage := &grpcStorage{id: uuid.New().String(), InmemoryStorage: kvstore.NewInmemoryStorage(), timeout: 1 * time.Second, logger: util.CreateLogger(logLevel)}
	for _, address := range addresses {
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		storage.clients = append(storage.clients, NewKVStoreClient(conn))
	}
	go storage.startServe(serverPort)
	return storage, nil
}

func (g *grpcStorage) startServe(port int) {
	listening, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("can't create listening to port %d: %v", port, err)
	}
	g.server = grpc.NewServer()
	RegisterKVStoreServer(g.server, g)
	g.logger.Info(fmt.Sprintf("kvstore %s is serving at %v", g.id, listening.Addr()))
	if err := g.server.Serve(listening); err != nil {
		log.Fatalf("kvstore %s: can't listen to port %d: %v", g.id, port, err)
	}
}

func (g *grpcStorage) StopServe() {
	g.logger.Info(fmt.Sprintf("stop serving kvstore: %s ", g.id))
	g.server.GracefulStop()
}

func (g *grpcStorage) StoreVal(ctx context.Context, storeValRequest *StoreValRequest) (*StoreValResponse, error) {
	var val interface{}
	err := json.Unmarshal([]byte(storeValRequest.Value), &val)
	if err != nil {
		return nil, err
	}

	err = g.InmemoryStorage.Put(storeValRequest.Storage, storeValRequest.Key, val)
	if err != nil {
		return nil, err
	}
	g.logger.Debug(fmt.Sprintf("grpc storage: %s : %d bytes stored in %s with key '%s' ", g.id, len(storeValRequest.Value), storeValRequest.Storage, storeValRequest.Key))
	return &StoreValResponse{}, nil
}

func (g *grpcStorage) Put(store, key string, val interface{}) error {
	json, err := json.Marshal(val)
	if err != nil {
		return err
	}

	for _, client := range g.clients {
		ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
		_, err := client.StoreVal(ctx, &StoreValRequest{Storage: store, Key: key, Value: string(json)})
		defer cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *grpcStorage) Shutdown() error {
	g.server.GracefulStop()
	return nil
}
