package model

import "text/template"

type MatchRequest struct {
	Template *template.Template `yaml:"-"`
	Scheme   string             `yaml:"scheme"`
	Host     string             `yaml:"host"`
	Method   string             `yaml:"method"`
	Path     string             `yaml:"path"`
	Query    map[string]string  `yaml:"query"`
	Headers  map[string]string  `yaml:"headers"`
}

type MockResponse struct {
	Template     *template.Template `yaml:"-"`
	StatusCode   int                `yaml:"statusCode"`
	Headers      map[string]string  `yaml:"headers"`
	Body         string             `yaml:"body"`
	BodyFileName string             `yaml:"bodyFileName"`
}

type MockEndpoint struct {
	Prio     int           `yaml:"prio"`
	Request  *MatchRequest `yaml:"request"`
	Response *MockResponse `yaml:"response"`
}

type Mock struct {
	Endpoints []*MockEndpoint `yaml:"endpoints" `
}
