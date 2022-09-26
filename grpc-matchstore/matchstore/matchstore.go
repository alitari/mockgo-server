package matchstore

import (
	context "context"
	"time"

	"github.com/alitari/mockgo/matches"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcMatchstore struct {
	*matches.InMemoryMatchstore
	client  MatchstoreClient
	timeout time.Duration
}

func NewGrpcMatchstore(addr string) (*GrpcMatchstore, error) {
	matchstore := &GrpcMatchstore{timeout: 1 * time.Second}
	if conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return nil, err
	} else {
		matchstore.client = NewMatchstoreClient(conn)
	}
	return matchstore, nil
}

func (g *GrpcMatchstore) GetMatches(endpointId string) ([]*matches.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	response, err := g.client.GetMatches(ctx, &EndPointRequest{Id: endpointId})
	defer cancel()
	if err != nil {
		return nil, err
	}
	res := []*matches.Match{}
	for _, match := range response.GetMatches() {
		res = append(res, mapMatch(match))
	}
	return res, nil
}

func mapMatch(grpcmatch *Match) *matches.Match {
	match := &matches.Match{EndpointId: grpcmatch.EndpointId, Timestamp: grpcmatch.Timestamp.AsTime(),
		ActualRequest:  &matches.ActualRequest{Method: grpcmatch.ActualRequest.Method, URL: grpcmatch.ActualRequest.Url, Header: mapHeader(grpcmatch.ActualRequest.Header), Host: grpcmatch.ActualRequest.Host},
		ActualResponse: &matches.ActualResponse{StatusCode: int(grpcmatch.ActualResponse.StatusCode), Header: grpcmatch.ActualResponse.Header}}
	return match
}

func mapHeader(header map[string]*HeaderValue) map[string][]string {
	res := map[string][]string{}
	for k, headerValue := range header {
		res[k] = append(res[k], headerValue.Val...)
	}
	return res
}
