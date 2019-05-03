package searchguard

import (
	"gopkg.in/yaml.v2"
)

//Roles are the roles for the ES Cluster
// root
//   roleName:
//     cluster:
//     expires:
//     indices:
//       indexName:
//         docType: [permissions]
type Roles map[string]Role

type Role struct {
	ClusterPermissions Permissions      `yaml:"cluster,omitempty"`
	ExpiresInMillis    int64            `yaml:"expires,omitempty"`
	IndicesPermissions IndexPermissions `yaml:"indices,omitempty"`
}
type Permissions []string

type IndexPermissions map[string]DocumentPermissions

type DocumentPermissions map[string]Permissions

func (roles *Roles) ToYaml() (string, error) {
	return toYaml(roles)
}
func (rolesmapping *RolesMapping) ToYaml() (string, error) {
	return toYaml(rolesmapping)
}
func toYaml(acl interface{}) (string, error) {
	var out []byte
	var err error
	if out, err = yaml.Marshal(acl); err != nil {
		return "", err
	}
	return string(out), nil
}

//SearchGuardRolesmapping are the mapping of username/groups to roles
// root
//  roleName
//    expires:
//    users:
//    groups:
type RolesMapping map[string]RoleMapping

type RoleMapping struct {
	ExpiresInMillis int64    `yaml:"expires,omitempty"`
	Users           []string `yaml:"users,omitempty"`
	Groups          []string `yaml:"groups,omitempty"`
}
