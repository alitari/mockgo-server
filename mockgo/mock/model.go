package mock

import (
	"regexp"
	"text/template"
)

/*
MatchRequest configuration model for a http request
*/
type MatchRequest struct {
	Host       string            `yaml:"host,omitempty" json:"host"`
	Method     string            `yaml:"method,omitempty" json:"method"`
	Path       string            `yaml:"path,omitempty" json:"path"`
	Query      map[string]string `yaml:"query,omitempty" json:"query"`
	Headers    map[string]string `yaml:"headers,omitempty" json:"headers"`
	Body       string            `yaml:"body,omitempty" json:"body"`
	BodyRegexp *regexp.Regexp    `yaml:"-" json:"-" `
}

/*
Response configuration model for a http response
*/
type Response struct {
	Template     *template.Template `yaml:"-" json:"-"`
	StatusCode   string             `yaml:"statusCode,omitempty" json:"statusCode"`
	Headers      string             `yaml:"headers,omitempty" json:"headers"`
	Body         string             `yaml:"body,omitempty" json:"body"`
	BodyFilename string             `yaml:"bodyFilename,omitempty" json:"bodyFilename"`
}

/*
Endpoint configuration model for a mock endpoint
*/
type Endpoint struct {
	ID       string        `yaml:"id" json:"id"`
	Mock     *Mock         `yaml:"-" json:"mock" `
	Prio     int           `yaml:"prio" json:"prio"`
	Request  *MatchRequest `yaml:"request" json:"request"`
	Response *Response     `yaml:"response" json:"response"`
}

/*
Mock configuration model for a mock file
*/
type Mock struct {
	Name      string      `yaml:"name" json:"name"`
	Endpoints []*Endpoint `yaml:"endpoints" json:"-"`
}

type epSearchNode struct {
	searchNodes   map[string]*epSearchNode
	endpoints     map[string][]*Endpoint
	pathParamName string
}
