package main

import (
	ext "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions"
	cl "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging"
)

// This file is intended to minimize the disruption of
// merging upstream to variants of OAuthProxy
func (proxy *OAuthProxy) registerExtensions(opts *ext.Options) {
	//fill in
	proxy.RegisterRequestHandlers(cl.NewHandlers(opts))
}
