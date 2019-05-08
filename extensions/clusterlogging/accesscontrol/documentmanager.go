package accesscontrol

import (
	"time"

	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/clients"
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/searchguard"
	cl "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/types"
)

type searchGuardClient interface {
	FetchRolesMapping() (*searchguard.RolesMapping, error)
	FetchRoles() (*searchguard.Roles, error)
	FlushRolesMapping(*searchguard.RolesMapping) error
	FlushRoles(*searchguard.Roles) error
}

//DocumentManager understands how to load and sync ACL documents
type DocumentManager struct {
	cl.ExtConfig
	sgclient searchGuardClient
}

//NewDocumentManager creates an instance or returns error
func NewDocumentManager(config cl.ExtConfig) (*DocumentManager, error) {
	sgClient, err := clients.NewSearchGuardClient()
	if err != nil {
		return nil, err
	}
	return &DocumentManager{
		config,
		sgClient,
	}, nil
}

//SyncACL to include the given UserInfo
func (dm *DocumentManager) SyncACL(userInfo *cl.UserInfo) error {
	if !dm.isInfraGroupMember(userInfo) {
		docs, err := dm.loadACL()
		if err != nil {
			return err
		}
		docs.ExpirePermissions()
		docs.AddUser(userInfo, nextExpireTime(dm.ExtConfig.PermissionExpirationMillis))
		if err = dm.writeACL(docs); err != nil {
			return err
		}
	}
	return nil
}
func (dm *DocumentManager) writeACL(docs *searchguard.ACLDocuments) error {
	if err := dm.sgclient.FlushRoles(&docs.Roles); err != nil {
		return err
	}
	if err := dm.sgclient.FlushRolesMapping(&docs.RolesMapping); err != nil {
		return err
	}
	return nil
}

func (dm *DocumentManager) loadACL() (*searchguard.ACLDocuments, error) {
	//TODO work on mget of roles/mappings
	roles, err := dm.sgclient.FetchRoles()
	if err != nil {
		return nil, err
	}
	rolesmapping, err := dm.sgclient.FetchRolesMapping()
	if err != nil {
		return nil, err
	}
	return &searchguard.ACLDocuments{
		Roles:        *roles,
		RolesMapping: *rolesmapping,
	}, nil
}

func (dm *DocumentManager) isInfraGroupMember(user *cl.UserInfo) bool {
	for _, group := range user.Groups {
		if group == dm.ExtConfig.InfraGroupName {
			return true
		}
	}
	return false
}

func nextExpireTime(expire int64) int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) + expire
}
