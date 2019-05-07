package clusterlogging

import (
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions"
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/clients"
	"k8s.io/apimachinery/pkg/util/sets"
)

// request
//   username
//   groups
//
//  flow ->
//    1. replace uri if kibana to kibana.usernamehash
//       replace content for index '.kibana' kibana profile changed
//    2. seed dashboards/index patterns
//    3. update sg
//       kibana role - infra: whitelisted group

type extension struct {
	openshiftCAs []string

	//whitelisted is the list of user and or serviceacccounts for which
	//all proxy logic is skipped (e.g. fluent)
	whitelisted sets.String
}

type requestContext struct {
	*UserInfo
}

//NewHandlers is the initializer for clusterlogging extensions
func NewHandlers(openshiftCAs []string) []extensions.RequestHandler {
	return []extensions.RequestHandler{
		&extension{
			openshiftCAs,
			sets.NewString(),
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
	context = &requestContext{
		userInfo,
	}

	return modRequest, nil
}

func (ext *extension) isWhiteListed(name string) bool {
	return ext.whitelisted.Has(name)
}

func userName(req *http.Request) string {
	return req.Header.Get("X-Forwarded-User")
}

func newUserInfo(ext *extension, req *http.Request) (*UserInfo, error) {
	projects, err := ext.fetchProjects(req.Header.Get("X-Forwarded-Access-Token"))
	if err != nil {
		return nil, err
	}
	info := &UserInfo{
		Username: userName(req),
		Projects: projects,
	}
	if groups, found := req.Header["X-Forwarded-Groups"]; found {
		info.Groups = groups
	}
	return info, nil
}

func (ex *extension) fetchProjects(token string) ([]Project, error) {
	client, err := clients.NewOpenShiftClient(ex.openshiftCAs, token)
	if err != nil {
		return nil, err
	}
	var json *simplejson.Json
	json, err = client.Get("/apis/project.openshift.io/v1/projects")
	if err != nil {
		return nil, err
	}
	projects := []Project{}
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
			projects = append(projects, Project{name, uid})
		}
	}
	return projects, nil
}

func (ex *extension) Name() string {
	return "addUserProjects"
}
