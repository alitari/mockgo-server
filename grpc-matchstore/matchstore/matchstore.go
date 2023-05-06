package matchstore

import (
	context "context"
	"fmt"
	"github.com/alitari/mockgo-server/mockgo/util"
	"go.uber.org/zap"
	"log"
	"net"
	"time"

	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcMatchstore struct {
	id string
	matches.Matchstore
	clients []MatchstoreClient
	timeout time.Duration
	logger  *zap.Logger
	UnimplementedMatchstoreServer
	server       *grpc.Server
	transferLock bool
}

/*
NewGrpcMatchstore creates a new distributed matches.Matchstore
*/
func NewGrpcMatchstore(addresses []string, serverPort int, capacity uint16, logLevel int) (matches.Matchstore, error) {
	matchstore := &grpcMatchstore{id: uuid.New().String(), Matchstore: matches.NewInMemoryMatchstore(capacity), timeout: 1 * time.Second, transferLock: false, logger: util.CreateLogger(logLevel)}
	for _, address := range addresses {
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		matchstore.clients = append(matchstore.clients, NewMatchstoreClient(conn))
	}
	go matchstore.startServe(serverPort)
	return matchstore, nil
}

func (g *grpcMatchstore) startServe(port int) {
	listening, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("can't create listening to port %d: %v", port, err)
	}
	g.server = grpc.NewServer()
	RegisterMatchstoreServer(g.server, g)
	g.logger.Info(fmt.Sprintf("matchstore %s is serving at %v", g.id, listening.Addr()))
	if err := g.server.Serve(listening); err != nil {
		log.Fatalf("can't listen to port %d: %v", port, err)
	}
}

func (g *grpcMatchstore) StopServe() {
	g.logger.Info(fmt.Sprintf("stop serving matchstore: %s ", g.id))
	g.server.GracefulStop()
}

func (g *grpcMatchstore) FetchMatches(ctx context.Context, endpointRequest *EndPointRequest) (*MatchesResponse, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : fetching matches for endpointId: %s ...", g.id, endpointRequest.Id))
	matches, err := g.Matchstore.GetMatches(endpointRequest.Id)
	if err != nil {
		return nil, err
	}
	protoMatches := []*Match{}
	for _, match := range matches {
		protoMatches = append(protoMatches, mapMatch(match))
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d matches", g.id, len(protoMatches)))
	return &MatchesResponse{Matches: protoMatches}, nil
}

func (g *grpcMatchstore) FetchMatchesCount(ctx context.Context, endpointRequest *EndPointRequest) (*MatchesCountResponse, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : fetching matches count for endpointId: %s ...", g.id, endpointRequest.Id))
	matchesCount, err := g.Matchstore.GetMatchesCount(endpointRequest.Id)
	if err != nil {
		return nil, err
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d matches", g.id, matchesCount))
	return &MatchesCountResponse{MatchesCount: matchesCount}, nil
}

func (g *grpcMatchstore) FetchMismatches(context.Context, *MismatchRequest) (*MismatchesResponse, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : fetching mismatches...", g.id))
	mismatches, err := g.Matchstore.GetMismatches()
	if err != nil {
		return nil, err
	}
	protoMismatches := []*Mismatch{}
	for _, mismatch := range mismatches {
		protoMismatches = append(protoMismatches, mapMismatch(mismatch))
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d mismatches", g.id, len(protoMismatches)))
	return &MismatchesResponse{Mismatches: protoMismatches}, nil
}

func (g *grpcMatchstore) FetchMismatchesCount(context.Context, *MismatchRequest) (*MismatchesCountResponse, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : fetching mismatches count ...", g.id))
	mismatchesCount, err := g.Matchstore.GetMismatchesCount()
	if err != nil {
		return nil, err
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d mismatches", g.id, mismatchesCount))
	return &MismatchesCountResponse{MismatchesCount: mismatchesCount}, nil
}

func (g *grpcMatchstore) RemoveMatches(ctx context.Context, endpointRequest *EndPointRequest) (*RemoveResponse, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : remove matches for endpoint %s ...", g.id, endpointRequest.Id))
	if err := g.Matchstore.DeleteMatches(endpointRequest.Id); err != nil {
		return nil, err
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : matches removed", g.id))
	return &RemoveResponse{}, nil
}

func (g *grpcMatchstore) RemoveMismatches(ctx context.Context, in *MismatchRequest) (*RemoveResponse, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : remove mismatches ...", g.id))
	if err := g.Matchstore.DeleteMismatches(); err != nil {
		return nil, err
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : mismatches removed", g.id))
	return &RemoveResponse{}, nil
}

func (g *grpcMatchstore) GetMatches(endpointID string) ([]*matches.Match, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : get matches for endpointId: %s ...", g.id, endpointID))
	matchesFromClients := [][]*matches.Match{}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for i, client := range g.clients {
		response, err := client.FetchMatches(ctx, &EndPointRequest{Id: endpointID})
		if err != nil {
			return nil, err
		}
		matchesFromClients = append(matchesFromClients, []*matches.Match{})
		for _, match := range response.GetMatches() {
			matchesFromClients[i] = append(matchesFromClients[i], mapProtoMatch(match))
		}
	}
	result := []*matches.Match{}
	finish := false
	pos := 0
	for !finish {
		finish = true
		for i := 0; i < len(g.clients); i++ {
			if pos < len(matchesFromClients[i]) {
				finish = false
				result = append(result, matchesFromClients[i][pos])
			}
		}
		pos++
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d matches as result for endpointId: %s ...", g.id, len(result), endpointID))
	return result, nil
}

func (g *grpcMatchstore) GetMatchesCount(endpointID string) (uint64, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : get matchesCount for endpointId: %s ...", g.id, endpointID))
	matchesCount := uint64(0)
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		response, err := client.FetchMatchesCount(ctx, &EndPointRequest{Id: endpointID})
		if err != nil {
			return uint64(0), err
		}
		matchesCount = matchesCount + uint64(response.MatchesCount)
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d matches as result for endpointId: %s", g.id, matchesCount, endpointID))
	return matchesCount, nil
}

func (g *grpcMatchstore) GetMismatches() ([]*matches.Mismatch, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : get mismatches ...", g.id))
	mismatches := []*matches.Mismatch{}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		response, err := client.FetchMismatches(ctx, &MismatchRequest{})
		if err != nil {
			return nil, err
		}
		for _, match := range response.GetMismatches() {
			mismatches = append(mismatches, mapProtoMismatch(match))
		}
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d mismatches as result", g.id, len(mismatches)))
	return mismatches, nil
}
func (g *grpcMatchstore) GetMismatchesCount() (uint64, error) {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : get mismatches count ...", g.id))
	mismatchesCount := uint64(0)
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		response, err := client.FetchMismatchesCount(ctx, &MismatchRequest{})
		if err != nil {
			return uint64(0), err
		}
		mismatchesCount = mismatchesCount + uint64(response.MismatchesCount)
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : return %d mismatches as result", g.id, mismatchesCount))
	return mismatchesCount, nil
}

func (g *grpcMatchstore) DeleteMatches(endpointID string) error {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : delete matches for endpointId: %s ...", g.id, endpointID))
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		_, err := client.RemoveMatches(ctx, &EndPointRequest{Id: endpointID})
		if err != nil {
			return err
		}
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : matches for endpointId: %s deleted", g.id, endpointID))
	return nil
}

func (g *grpcMatchstore) DeleteMismatches() error {
	g.logger.Debug(fmt.Sprintf("matchstore: %s : delete mismatches ...", g.id))
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		_, err := client.RemoveMismatches(ctx, &MismatchRequest{})
		if err != nil {
			return err
		}
	}
	g.logger.Debug(fmt.Sprintf("matchstore: %s : mismatches deleted", g.id))
	return nil
}

func mapProtoMatch(protomatch *Match) *matches.Match {
	match := &matches.Match{EndpointID: protomatch.EndpointId, Timestamp: protomatch.Timestamp.AsTime(),
		ActualRequest:  &matches.ActualRequest{Method: protomatch.ActualRequest.Method, URL: protomatch.ActualRequest.Url, Header: mapProtoHeader(protomatch.ActualRequest.Header), Host: protomatch.ActualRequest.Host},
		ActualResponse: &matches.ActualResponse{StatusCode: int(protomatch.ActualResponse.StatusCode), Header: mapProtoHeader(protomatch.ActualResponse.Header)}}
	return match
}

func mapMatch(match *matches.Match) *Match {
	protoMatch := &Match{EndpointId: match.EndpointID, Timestamp: timestamppb.New(match.Timestamp),
		ActualRequest:  &ActualRequest{Method: match.ActualRequest.Method, Url: match.ActualRequest.URL, Header: mapHeader(match.ActualRequest.Header), Host: match.ActualRequest.Host},
		ActualResponse: &ActualResponse{StatusCode: int32(match.ActualResponse.StatusCode), Header: mapHeader(match.ActualResponse.Header)}}
	return protoMatch
}

func mapProtoMismatch(protomismatch *Mismatch) *matches.Mismatch {
	mismatch := &matches.Mismatch{MismatchDetails: protomismatch.MismatchDetails, Timestamp: protomismatch.Timestamp.AsTime(),
		ActualRequest: &matches.ActualRequest{Method: protomismatch.ActualRequest.Method, URL: protomismatch.ActualRequest.Url, Header: mapProtoHeader(protomismatch.ActualRequest.Header), Host: protomismatch.ActualRequest.Host}}
	return mismatch
}

func mapMismatch(mismatch *matches.Mismatch) *Mismatch {
	protoMismatch := &Mismatch{MismatchDetails: mismatch.MismatchDetails, Timestamp: timestamppb.New(mismatch.Timestamp),
		ActualRequest: &ActualRequest{Method: mismatch.ActualRequest.Method, Url: mismatch.ActualRequest.URL, Header: mapHeader(mismatch.ActualRequest.Header), Host: mismatch.ActualRequest.Host},
	}
	return protoMismatch
}

func mapProtoHeader(header map[string]*HeaderValue) map[string][]string {
	res := map[string][]string{}
	for k, headerValue := range header {
		res[k] = headerValue.Val
	}
	return res
}

func mapHeader(header map[string][]string) map[string]*HeaderValue {
	res := map[string]*HeaderValue{}
	for k, val := range header {
		res[k] = &HeaderValue{Val: val}
	}
	return res
}
