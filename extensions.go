package main

import (
	cl "github.com/openshift/elasticsearch-cluster-logging-proxy/extensions/clusterlogging"
)

// This file is intended to minimize the disruption of
// merging upstream to variants of OAuthProxy
func (proxy *OAuthProxy) registerExtensions() {
	//fill in
	proxy.RegisterRequestHandlers(cl.NewHandlers())
}
