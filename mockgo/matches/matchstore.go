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
	StatusCode int               `json:"statusCode"`
	Header     map[string][]string `json:"header"`
}

type Matchstore interface {
	HasMatchesCountOnly() bool
	HasMismatchesCountOnly() bool
	GetMatches(endpointId string) ([]*Match, error)
	AddMatches(matches map[string][]*Match) error
	GetMatchesCount(endpointId string) (int64, error)
	AddMatchesCount(matchesCount map[string]int64) error
	GetMismatches() ([]*Mismatch, error)
	AddMismatches([]*Mismatch) error
	GetMismatchesCount() (int64, error)
	AddMismatchesCount(int64) error
	DeleteMatches() error
	DeleteMismatches() error
}