package searchguard

import (
	"encoding/json"

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
	ClusterPermissions Permissions      `yaml:"cluster,omitempty" json:"cluster,omitempty"`
	ExpiresInMillis    int64            `yaml:"expires,omitempty" json:"expires,omitempty"`
	IndicesPermissions IndexPermissions `yaml:"indices,omitempty" json:"indices,omitempty"`
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

func (roles *Roles) FromJson(acl string) error {
	if err := json.Unmarshal([]byte(acl), roles); err != nil {
		return err
	}
	return nil
}

func (rolesmapping *RolesMapping) FromJson(acl string) error {
	if err := json.Unmarshal([]byte(acl), rolesmapping); err != nil {
		return err
	}
	return nil
}

//Rolesmapping are the mapping of username/groups to roles
// root
//  roleName
//    expires:
//    users:
//    groups:
type RolesMapping map[string]RoleMapping

type RoleMapping struct {
	ExpiresInMillis int64    `yaml:"expires,omitempty" json:"expires,omitempty"`
	Users           []string `yaml:"users,omitempty" json:"users,omitempty"`
	Groups          []string `yaml:"groups,omitempty" yaml:"groups,omitempty"`
}
