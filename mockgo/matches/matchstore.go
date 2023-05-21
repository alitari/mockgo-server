package matches

import "time"

/*
Match datamodel for a http request which hit an endpoint
*/
type Match struct {
	EndpointID     string          `json:"endpointId"`
	Timestamp      time.Time       `json:"timestamp"`
	ActualRequest  *ActualRequest  `json:"actualRequest"`
	ActualResponse *ActualResponse `json:"actualResponse"`
}

/*
Mismatch datamodel for a http request which missed an endpoint
*/
type Mismatch struct {
	MismatchDetails string         `json:"MismatchDetails"`
	Timestamp       time.Time      `json:"timestamp"`
	ActualRequest   *ActualRequest `json:"actualRequest"`
}

/*
ActualRequest datamodel for an incoming http request which is stored for a match or mismatch
*/
type ActualRequest struct {
	Method string              `json:"method" `
	URL    string              `json:"url" `
	Header map[string][]string `json:"header" `
	Host   string              `json:"host" `
}

/*
ActualResponse datamodel for an outgoing http response from a request which is stored for a match
*/
type ActualResponse struct {
	StatusCode int                 `json:"statusCode"`
	Header     map[string][]string `json:"header"`
}

/*
Matchstore is the interface for a storage which holds the http requests which matches mock endpoints.
*/
type Matchstore interface {
	GetMatches(endpointID string) ([]*Match, error)
	GetMatchesCount(endpointID string) (uint64, error)
	GetMismatches() ([]*Mismatch, error)
	AddMatch(endpointID string, match *Match) error
	AddMismatch(*Mismatch) error
	GetMismatchesCount() (uint64, error)
	DeleteMatches(endpointID string) error
	DeleteMismatches() error
	Shutdown() error
}
