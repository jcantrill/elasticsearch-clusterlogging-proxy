package config

import (
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions"
)

//ExtConfig defines configuration the proxy may use to make
//decisions (e.g. role name)
type ExtConfig struct {
	KibanaIndexMode
	//InfraGroupName is the groupname for which a user should be considered an
	//administrator and will be granted the ocp_admin_role
	InfraGroupName string

	//PermissionExpirationMillis  the time when permissions expire
	PermissionExpirationMillis int64

	//Options passed to the proxy
	Options extensions.Options
}

//KibanaIndexMode is the mode the proxy uses to generate a user's kibana index
type KibanaIndexMode string

const (
	//KibanaIndexModeSharedOps all users of the InfraGroupName will share a common Kibana index
	KibanaIndexModeSharedOps KibanaIndexMode = "sharedOps"
)
