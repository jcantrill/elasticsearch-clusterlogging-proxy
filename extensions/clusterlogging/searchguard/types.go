package searchguard

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strings"

	cl "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/types"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ACLDocuments struct {
	Roles
	RolesMapping
}

type Serializable interface {
	ToJson() (string, error)
	Type() string
}

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

func (roles *Roles) Type() string {
	return "roles"
}
func (rolesmapping *RolesMapping) Type() string {
	return "rolesmapping"
}
func (roles *Roles) ToYaml() (string, error) {
	return toYaml(roles)
}
func (rolesmapping *RolesMapping) ToYaml() (string, error) {
	return toYaml(rolesmapping)
}

func (roles *Roles) ToJson() (string, error) {
	return ToJson(roles)
}

func (rolesmapping *RolesMapping) ToJson() (string, error) {
	return ToJson(rolesmapping)
}

func toYaml(acl interface{}) (string, error) {
	var out []byte
	var err error
	if out, err = yaml.Marshal(acl); err != nil {
		return "", err
	}
	return string(out), nil
}

func ToJson(acl interface{}) (string, error) {
	log.Tracef("Converting acl to json: %+v", acl)
	var out []byte
	var err error
	if out, err = json.Marshal(acl); err != nil {
		return "", err
	}
	resp := string(out)
	log.Tracef("Converted: %s", resp)
	return resp, nil
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

//AddUser permissions to the ACL documents
func (docs *ACLDocuments) AddUser(user *cl.UserInfo, expires int64) {
	roleName := roleName(user)
	docs.Roles[roleName] = Role{
		ClusterPermissions: Permissions{"CLUSTER_MONITOR_KIBANA", "USER_CLUSTER_OPERATIONS"},
		ExpiresInMillis:    expires,
		IndicesPermissions: newSearchGuardDocumentPermissions(user),
	}
	docs.RolesMapping[roleName] = RoleMapping{
		ExpiresInMillis: expires,
		Users:           []string{user.Username},
		Groups:          user.Groups,
	}
}

func (docs *ACLDocuments) ExpirePermissions() {
}

func newSearchGuardDocumentPermissions(user *cl.UserInfo) IndexPermissions {
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

func kibanaIndexName(user *cl.UserInfo) string {
	return fmt.Sprintf(".kibana.%s", usernameHash(user))
}

func roleName(user *cl.UserInfo) string {
	return fmt.Sprintf("gen_user_%s", usernameHash(user))
}

func usernameHash(user *cl.UserInfo) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(user.Username)))
}
