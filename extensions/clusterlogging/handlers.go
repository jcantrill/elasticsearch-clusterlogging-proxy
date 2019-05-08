package clusterlogging

import (
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions"
	ac "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/accesscontrol"
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/clients"
	config "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

type extension struct {
	openshiftCAs []string

	//whitelisted is the list of user and or serviceacccounts for which
	//all proxy logic is skipped (e.g. fluent)
	whitelisted     sets.String
	documentManager *ac.DocumentManager
}

type requestContext struct {
	*config.UserInfo
}

//NewHandlers is the initializer for clusterlogging extensions
func NewHandlers(openshiftCAs []string) []extensions.RequestHandler {
	config := config.ExtConfig{
		KibanaIndexMode:            config.KibanaIndexModeSharedOps,
		InfraGroupName:             "system:cluster-admins",
		PermissionExpirationMillis: 1000 * 2, //2 minutes
	}
	dm, err := ac.NewDocumentManager(config)
	if err != nil {
		log.Panicf("Unable to initialize the cluster logging proxy extension %v", err)
	}
	return []extensions.RequestHandler{
		&extension{
			openshiftCAs,
			sets.NewString(),
			dm,
		},
	}
}

func (ext *extension) Process(req *http.Request, context interface{}) (*http.Request, error) {
	if ext.isWhiteListed(userName(req)) {
		return req, nil
	}
	modRequest := req
	userInfo, err := newUserInfo(ext, req)
	if err != nil {
		return req, err
	}
	// modify kibana request
	// seed kibana dashboards
	if err = ext.documentManager.SyncACL(userInfo); err != nil {
		return nil, err
	}

	return modRequest, nil
}

func (ext *extension) isWhiteListed(name string) bool {
	return ext.whitelisted.Has(name)
}

func userName(req *http.Request) string {
	return req.Header.Get("X-Forwarded-User")
}

func newUserInfo(ext *extension, req *http.Request) (*config.UserInfo, error) {
	projects, err := ext.fetchProjects(req.Header.Get("X-Forwarded-Access-Token"))
	if err != nil {
		return nil, err
	}
	info := &config.UserInfo{
		Username: userName(req),
		Projects: projects,
	}
	if groups, found := req.Header["X-Forwarded-Groups"]; found {
		info.Groups = groups
	}
	return info, nil
}

func (ext *extension) fetchProjects(token string) ([]config.Project, error) {
	client, err := clients.NewOpenShiftClient(ext.openshiftCAs, token)
	if err != nil {
		return nil, err
	}
	var json *simplejson.Json
	json, err = client.Get("/apis/project.openshift.io/v1/projects")
	if err != nil {
		return nil, err
	}
	projects := []config.Project{}
	if items, ok := json.CheckGet("items"); ok {
		total := len(items.MustArray())
		for i := 0; i < total; i++ {
			//check for missing?
			var name, uid string
			if value := items.GetIndex(i).GetPath("metadata", "name"); value.Interface() != nil {
				name = value.MustString()
			}
			if value := items.GetIndex(i).GetPath("metadata", "uid"); value.Interface() != nil {
				uid = value.MustString()
			}
			projects = append(projects, config.Project{Name: name, UUID: uid})
		}
	}
	return projects, nil
}

func (ext *extension) Name() string {
	return "addUserProjects"
}
