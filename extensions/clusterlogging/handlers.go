package clusterlogging

import (
	"net/http"

	"github.com/bitly/go-simplejson"

	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions"
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
}

//NewHandlers is the initializer for clusterlogging extensions
func NewHandlers(openshiftCAs []string) []extensions.RequestHandler {
	return []extensions.RequestHandler{
		&extension{openshiftCAs},
	}
}

func (ex *extension) Process(req *http.Request) (*http.Request, error) {
	token := req.Header.Get("X-Forwarded-Access-Token")
	client, err := newOpenShiftClient(ex.openshiftCAs, token)
	if err != nil {
		return nil, err
	}
	var json *simplejson.Json
	json, err = client.Get("/apis/project.openshift.io/v1/projects")
	if err != nil {
		return nil, err
	}
	projects := []string{}
	if items, ok := json.CheckGet("items"); ok {
		total := len(items.MustArray())
		for i := 0; i < total; i++ {
			if name := items.GetIndex(i).GetPath("metadata", "name"); name.Interface() != nil {
				projects = append(projects, name.MustString())
			}
		}
	}
	modRequest := req
	modRequest.Header["X-Forwarded-Projects"] = projects
	return modRequest, nil
}

func (ex *extension) Name() string {
	return "addUserProjects"
}
