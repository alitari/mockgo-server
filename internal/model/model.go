package model

import "text/template"

type Request struct {
	Template *template.Template
	Scheme   string            `json:"scheme" yaml:"scheme"`
	Host     string            `json:"host" yaml:"host"`
	Method   string            `json:"method" yaml:"method"`
	Path     string            `json:"path" yaml:"path"`
	Query    map[string]string `json:"query" yaml:"query"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
}

type Response struct {
	Template *template.Template
	StatusCode   int               `json:"statusCode" yaml:"statusCode"`
	Headers      map[string]string `json:"headers" yaml:"headers"`
	Body         string            `json:"body" yaml:"body"`
	BodyFileName string            `json:"bodyFileName" yaml:"bodyFileName"`
}

type Endpoint struct {
	Prio     int      `json:"prio" yaml:"prio"`
	Request  *Request  `json:"request" yaml:"request"`
	Response *Response `json:"response" yaml:"response"`
}

type Mock struct {
	Endpoints []*Endpoint `json:"endpoints" yaml:"endpoints" `
}
