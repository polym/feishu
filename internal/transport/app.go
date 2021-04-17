package transport

import (
	"io"
)

type AuthKind int

const (
	AuthKindNone = iota
	AuthKindTenant
)

type Request struct {
	Method      string
	Url         string
	Params      map[string]string
	ContentType string
	BodySpec    interface{}
	Body        io.Reader
}

type Transport interface {
	Do(req *Request, resp interface{}, kind AuthKind) error
}

type EmbedData interface {
	Embed()
}

type LabelEmbedData struct{}

func (e LabelEmbedData) Embed() {}
