package model

import "text/template"

type Request struct {
	Template *template.Template `yaml:"-"`
	Scheme   string             `yaml:"scheme"`
	Host     string             `yaml:"host"`
	Method   string             `yaml:"method"`
	Path     string             `yaml:"path"`
	Query    map[string]string  `yaml:"query"`
	Headers  map[string]string  `yaml:"headers"`
}

type Response struct {
	Template     *template.Template `yaml:"-"`
	StatusCode   int                `yaml:"statusCode"`
	Headers      map[string]string  `yaml:"headers"`
	Body         string             `yaml:"body"`
	BodyFileName string             `yaml:"bodyFileName"`
}

type Endpoint struct {
	Prio     int       `yaml:"prio"`
	Request  *Request  `yaml:"request"`
	Response *Response `yaml:"response"`
}

type Mock struct {
	Endpoints []*Endpoint `yaml:"endpoints" `
}
