package matches

import "time"

type Match struct {
	EndpointId     string          `json:"endpointId"`
	Timestamp      time.Time       `json:"timestamp"`
	ActualRequest  *ActualRequest  `json:"actualRequest"`
	ActualResponse *ActualResponse `json:"actualResponse"`
}

type Mismatch struct {
	MismatchDetails string         `json:"MismatchDetails"`
	Timestamp       time.Time      `json:"timestamp"`
	ActualRequest   *ActualRequest `json:"actualRequest"`
}

type ActualRequest struct {
	Method string              `json:"method" `
	URL    string              `json:"url" `
	Header map[string][]string `json:"header" `
	Host   string              `json:"host" `
}

type ActualResponse struct {
	StatusCode int                 `json:"statusCode"`
	Header     map[string][]string `json:"header"`
}

type Matchstore interface {
	GetMatches(endpointId string) ([]*Match, error)
	GetMatchesCount(endpointId string) (uint64, error)
	GetMismatches() ([]*Mismatch, error)
	AddMatch(endpointId string, match *Match) error
	AddMismatch(*Mismatch) error
	GetMismatchesCount() (uint64, error)
	DeleteMatches(endpointId string) error
	DeleteMismatches() error
}
