package mock

import (
	"regexp"
	"text/template"
)

type MatchRequest struct {
	Scheme     string            `yaml:"scheme" json:"scheme"`
	Host       string            `yaml:"host" json:"host"`
	Method     string            `yaml:"method" json:"method"`
	Path       string            `yaml:"path" json:"path"`
	Query      map[string]string `yaml:"query" json:"query"`
	Headers    map[string]string `yaml:"headers" json:"headers"`
	Body       string            `yaml:"body" json:"body"`
	BodyRegexp *regexp.Regexp    `yaml:"-" json:"-" `
}

type MockResponse struct {
	Template     *template.Template `yaml:"-" json:"-"`
	StatusCode   string             `yaml:"statusCode" json:"statusCode"`
	Headers      string             `yaml:"headers" json:"headers"`
	Body         string             `yaml:"body" json:"body"`
	BodyFilename string             `yaml:"bodyFilename" json:"bodyFilename"`
}

type MockEndpoint struct {
	Id       string        `yaml:"id" json:"id"`
	Mock     *Mock         `yaml:"-" json:"mock" `
	Prio     int           `yaml:"prio" json:"prio"`
	Request  *MatchRequest `yaml:"request" json:"request"`
	Response *MockResponse `yaml:"response" json:"response"`
}

type Mock struct {
	Name      string          `yaml:"name" json:"name"`
	Endpoints []*MockEndpoint `yaml:"endpoints" json:"-"`
}

type EpSearchNode struct {
	SearchNodes   map[string]*EpSearchNode
	Endpoints     map[string][]*MockEndpoint
	PathParamName string
}