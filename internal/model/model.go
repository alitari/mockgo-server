package model

type Request struct {
	Scheme  string            `json:"scheme"`
	Host    string            `json:"host"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   map[string]string `json:"query"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	StatusCode   int               `json:"statusCode"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	BodyFileName string            `json:"bodyFileName"`
}

type Endpoint struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Mock struct {
	Description string     `json:"description"`
	Endpoints   []Endpoint `json:"endpoints"`
}
