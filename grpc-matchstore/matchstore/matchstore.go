package matchstore

import (
	context "context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcMatchstore struct {
	id string
	*matches.InMemoryMatchstore
	clients []MatchstoreClient
	timeout time.Duration
	logger  *logging.LoggerUtil
	UnimplementedMatchstoreServer
	server       *grpc.Server
	transferLock bool
}

func NewGrpcMatchstore(addresses []string, serverPort int, capacity uint16, logger *logging.LoggerUtil) (*GrpcMatchstore, error) {
	matchstore := &GrpcMatchstore{id: uuid.New().String(), InMemoryMatchstore: matches.NewInMemoryMatchstore(capacity), timeout: 1 * time.Second, transferLock: false, logger: logger}
	for _, address := range addresses {
		if conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
			return nil, err
		} else {
			matchstore.clients = append(matchstore.clients, NewMatchstoreClient(conn))
		}
	}
	go matchstore.startServe(serverPort)

	return matchstore, nil
}

func (g *GrpcMatchstore) startServe(port int) {
	listening, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("can't create listening to port %d: %v", port, err)
	}
	g.server = grpc.NewServer()
	RegisterMatchstoreServer(g.server, g)
	g.logger.LogWhenVerbose(fmt.Sprintf("matchstore %s is serving at %v", g.id, listening.Addr()))
	if err := g.server.Serve(listening); err != nil {
		log.Fatalf("can't listen to port %d: %v", port, err)
	}
}

func (g *GrpcMatchstore) StopServe() {
	g.logger.LogWhenVerbose(fmt.Sprintf("stop serving matchstore: %s ", g.id))
	g.server.GracefulStop()
}

func (g *GrpcMatchstore) FetchMatches(ctx context.Context, endpointRequest *EndPointRequest) (*MatchesResponse, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : fetching matches for endpointId: %s ...", g.id, endpointRequest.Id))
	matches, err := g.InMemoryMatchstore.GetMatches(endpointRequest.Id)
	if err != nil {
		return nil, err
	}
	protoMatches := []*Match{}
	for _, match := range matches {
		protoMatches = append(protoMatches, mapMatch(match))
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d matches", g.id, len(protoMatches)))
	return &MatchesResponse{Matches: protoMatches}, nil
}

func (g *GrpcMatchstore) FetchMatchesCount(ctx context.Context, endpointRequest *EndPointRequest) (*MatchesCountResponse, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : fetching matches count for endpointId: %s ...", g.id, endpointRequest.Id))
	matchesCount, err := g.InMemoryMatchstore.GetMatchesCount(endpointRequest.Id)
	if err != nil {
		return nil, err
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d matches", g.id, matchesCount))
	return &MatchesCountResponse{MatchesCount: matchesCount}, nil
}

func (g *GrpcMatchstore) FetchMismatches(context.Context, *MismatchRequest) (*MismatchesResponse, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : fetching mismatches...", g.id))
	mismatches, err := g.InMemoryMatchstore.GetMismatches()
	if err != nil {
		return nil, err
	}
	protoMismatches := []*Mismatch{}
	for _, mismatch := range mismatches {
		protoMismatches = append(protoMismatches, mapMismatch(mismatch))
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d mismatches", g.id, len(protoMismatches)))
	return &MismatchesResponse{Mismatches: protoMismatches}, nil
}

func (g *GrpcMatchstore) FetchMismatchesCount(context.Context, *MismatchRequest) (*MismatchesCountResponse, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : fetching mismatches count ...", g.id))
	mismatchesCount, err := g.InMemoryMatchstore.GetMismatchesCount()
	if err != nil {
		return nil, err
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d mismatches", g.id, mismatchesCount))
	return &MismatchesCountResponse{MismatchesCount: mismatchesCount}, nil
}

func (g *GrpcMatchstore) RemoveMatches(ctx context.Context, endpointRequest *EndPointRequest) (*RemoveResponse, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : remove matches for endpoint %s ...", g.id, endpointRequest.Id))
	if err := g.InMemoryMatchstore.DeleteMatches(endpointRequest.Id); err != nil {
		return nil, err
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : matches removed", g.id))
	return &RemoveResponse{}, nil
}

func (g *GrpcMatchstore) RemoveMismatches(ctx context.Context, in *MismatchRequest) (*RemoveResponse, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : remove mismatches ...", g.id))
	if err := g.InMemoryMatchstore.DeleteMismatches(); err != nil {
		return nil, err
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : mismatches removed", g.id))
	return &RemoveResponse{}, nil
}

func (g *GrpcMatchstore) GetMatches(endpointId string) ([]*matches.Match, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : get matches for endpointId: %s ...", g.id, endpointId))
	matchesFromClients := [][]*matches.Match{}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for i, client := range g.clients {
		response, err := client.FetchMatches(ctx, &EndPointRequest{Id: endpointId})
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
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d matches as result for endpointId: %s ...", g.id, len(result), endpointId))
	return result, nil
}

func (g *GrpcMatchstore) GetMatchesCount(endpointId string) (uint64, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : get matchesCount for endpointId: %s ...", g.id, endpointId))
	matchesCount := uint64(0)
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		response, err := client.FetchMatchesCount(ctx, &EndPointRequest{Id: endpointId})
		if err != nil {
			return uint64(0), err
		}
		matchesCount = matchesCount + uint64(response.MatchesCount)
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d matches as result for endpointId: %s", g.id, matchesCount, endpointId))
	return matchesCount, nil
}

func (g *GrpcMatchstore) GetMismatches() ([]*matches.Mismatch, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : get mismatches ...", g.id))
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
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d mismatches as result", g.id, len(mismatches)))
	return mismatches, nil
}
func (g *GrpcMatchstore) GetMismatchesCount() (uint64, error) {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : get mismatches count ...", g.id))
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
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : return %d mismatches as result", g.id, mismatchesCount))
	return mismatchesCount, nil
}

func (g *GrpcMatchstore) DeleteMatches(endpointId string) error {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : delete matches for endpointId: %s ...", g.id, endpointId))
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		_, err := client.RemoveMatches(ctx, &EndPointRequest{Id: endpointId})
		if err != nil {
			return err
		}
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : matches for endpointId: %s deleted", g.id, endpointId))
	return nil
}

func (g *GrpcMatchstore) DeleteMismatches() error {
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : delete mismatches ...", g.id))
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	for _, client := range g.clients {
		_, err := client.RemoveMismatches(ctx, &MismatchRequest{})
		if err != nil {
			return err
		}
	}
	g.logger.LogWhenDebug(fmt.Sprintf("matchstore: %s : mismatches deleted", g.id))
	return nil
}

func mapProtoMatch(protomatch *Match) *matches.Match {
	match := &matches.Match{EndpointId: protomatch.EndpointId, Timestamp: protomatch.Timestamp.AsTime(),
		ActualRequest:  &matches.ActualRequest{Method: protomatch.ActualRequest.Method, URL: protomatch.ActualRequest.Url, Header: mapProtoHeader(protomatch.ActualRequest.Header), Host: protomatch.ActualRequest.Host},
		ActualResponse: &matches.ActualResponse{StatusCode: int(protomatch.ActualResponse.StatusCode), Header: mapProtoHeader(protomatch.ActualResponse.Header)}}
	return match
}

func mapMatch(match *matches.Match) *Match {
	protoMatch := &Match{EndpointId: match.EndpointId, Timestamp: timestamppb.New(match.Timestamp),
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
