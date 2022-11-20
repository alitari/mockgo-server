package kvstore

import (
	context "context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcKVStore struct {
	id string
	*kvstore.InmemoryKVStore
	clients []KVStoreClient
	timeout time.Duration
	logger  *logging.LoggerUtil
	UnimplementedKVStoreServer
	server *grpc.Server
}

func NewGrpcKVstore(addresses []string, serverPort int, logger *logging.LoggerUtil) (*GrpcKVStore, error) {
	kvstore := &GrpcKVStore{id: uuid.New().String(), InmemoryKVStore: kvstore.NewInmemoryKVStore(), timeout: 1 * time.Second, logger: logger}
	for _, address := range addresses {
		if conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
			return nil, err
		} else {
			kvstore.clients = append(kvstore.clients, NewKVStoreClient(conn))
		}
	}
	go kvstore.startServe(serverPort)

	return kvstore, nil
}

func (g *GrpcKVStore) startServe(port int) {
	listening, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("can't create listening to port %d: %v", port, err)
	}
	g.server = grpc.NewServer()
	RegisterKVStoreServer(g.server, g)
	g.logger.LogWhenVerbose(fmt.Sprintf("kvstore %s is serving at %v", g.id, listening.Addr()))
	if err := g.server.Serve(listening); err != nil {
		log.Fatalf("kvstore %s: can't listen to port %d: %v", g.id, port, err)
	}
}

func (g *GrpcKVStore) StopServe() {
	g.logger.LogWhenVerbose(fmt.Sprintf("stop serving kvstore: %s ", g.id))
	g.server.GracefulStop()
}

func (g *GrpcKVStore) StoreVal(ctx context.Context, storeValRequest *StoreValRequest) (*StoreValResponse, error) {
	var val interface{}
	err := json.Unmarshal([]byte(storeValRequest.Value), &val)
	if err != nil {
		return nil, err
	}

	err = g.InmemoryKVStore.PutVal(storeValRequest.Key, val)
	if err != nil {
		return nil, err
	}
	g.logger.LogWhenDebug(fmt.Sprintf("kvstore: %s : %d bytes stored with key '%s'", g.id, len(storeValRequest.Value), storeValRequest.Key))
	return &StoreValResponse{}, nil
}

func (g *GrpcKVStore) PutVal(key string, storeVal interface{}) error {
	json, err := json.Marshal(storeVal)
	if err != nil {
		return err
	}

	for _, client := range g.clients {
		ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
		_, err := client.StoreVal(ctx, &StoreValRequest{Key: key, Value: string(json)})
		defer cancel()
		if err != nil {
			return err
		}
	}
	return nil
}
