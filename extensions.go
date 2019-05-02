package main

import (
	cl "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging"
)

// This file is intended to minimize the disruption of
// merging upstream to variants of OAuthProxy
func (proxy *OAuthProxy) registerExtensions(openshiftCAs StringArray) {
	//fill in
	proxy.RegisterRequestHandlers(cl.NewHandlers(openshiftCAs))
}
