package clusterlogging

import (
	"net/http"

	"github.com/openshift/elasticsearch-cluster-logging-proxy/extensions"
)

func NewHandlers() []extensions.RequestHandler {
	return []extensions.RequestHandler{
		extensions.NewRequestHandler("appendHeader", AppendHeader),
	}
}

func AppendHeader(req *http.Request) (*http.Request, error) {
	req.Header.Add("foo", "bar")
	return req, nil
}
