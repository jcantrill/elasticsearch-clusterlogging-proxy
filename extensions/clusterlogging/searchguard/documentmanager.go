package searchguard

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	cl "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging"
)

type DocumentManager struct {
	cl.ExtConfig
	Roles
	RolesMapping
	nextExpireTime func(int64) int64
}

func NewDocumentManager(config cl.ExtConfig) *DocumentManager {
	return &DocumentManager{
		config,
		Roles{},
		RolesMapping{},
		nextExpireTime,
	}
}

func (dm *DocumentManager) isInfraGroupMember(user cl.UserInfo) bool {
	for _, group := range user.Groups {
		if group == dm.ExtConfig.InfraGroupName {
			return true
		}
	}
	return false
}

func (dm *DocumentManager) AddUser(user cl.UserInfo) {
	if dm.isInfraGroupMember(user) {
		return
	}
	roleName := roleName(user)
	expires := dm.nextExpireTime(dm.ExtConfig.PermissionExpirationMillis)
	dm.Roles[roleName] = Role{
		ClusterPermissions: Permissions{"CLUSTER_MONITOR_KIBANA", "USER_CLUSTER_OPERATIONS"},
		ExpiresInMillis:    expires,
		IndicesPermissions: newSearchGuardDocumentPermissions(user),
	}
	dm.RolesMapping[roleName] = RoleMapping{
		ExpiresInMillis: expires,
		Users:           []string{user.Username},
		Groups:          user.Groups,
	}
}

func newSearchGuardDocumentPermissions(user cl.UserInfo) IndexPermissions {
	permissions := IndexPermissions{}
	permissions[fix(kibanaIndexName(user))] = DocumentPermissions{
		"*": Permissions{
			"INDEX_KIBANA",
		},
	}
	for _, project := range user.Projects {
		permissions[fix(projectIndexName(project))] = DocumentPermissions{
			"*": Permissions{
				"INDEX_PROJECT",
			},
		}
	}
	return permissions
}

func fix(indexName string) string {
	return strings.Replace(indexName, ".", "?", -1)
}

func projectIndexName(p cl.Project) string {
	return fmt.Sprintf("project.%s.%s.*", p.Name, p.UUID)
}

func kibanaIndexName(user cl.UserInfo) string {
	return fmt.Sprintf(".kibana.%s", usernameHash(user))
}

func roleName(user cl.UserInfo) string {
	return fmt.Sprintf("gen_user_%s", usernameHash(user))
}

func nextExpireTime(expire int64) int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) + expire
}

func usernameHash(user cl.UserInfo) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(user.Username)))
}
