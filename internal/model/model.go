package model

import "text/template"

type MatchRequest struct {
	Scheme  string            `yaml:"scheme"`
	Host    string            `yaml:"host"`
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Query   map[string]string `yaml:"query"`
	Headers map[string]string `yaml:"headers"`
}

type MockResponse struct {
	Template     *template.Template `yaml:"-"`
	StatusCode   int                `yaml:"statusCode"`
	Headers      map[string]string  `yaml:"headers"`
	Body         string             `yaml:"body"`
	BodyFilename string             `yaml:"bodyFilename"`
}

type MockEndpoint struct {
	Id       string        `yaml:"id"`
	Mock     *Mock         `yaml:"-"`
	Prio     int           `yaml:"prio"`
	Request  *MatchRequest `yaml:"request"`
	Response *MockResponse `yaml:"response"`
}

type Mock struct {
	Name      string          `yaml:"name"`
	Endpoints []*MockEndpoint `yaml:"endpoints" `
}
