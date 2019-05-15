package clients

import (
	"encoding/base64"
	"encoding/json"

	"github.com/oliveagle/jsonpath"

	ext "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions"
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/searchguard"
)

var (
	rolesLookup, _  = jsonpath.Compile("$._source.roles")
	rolesMapping, _ = jsonpath.Compile("$._source.rolesmapping")
)

type SearchGuardClient struct {
	esClient *ElasticsearchClient
}

func NewSearchGuardClient(opts ext.Options) (*SearchGuardClient, error) {
	esClient, err := NewElasticsearchClient(opts.SSLInsecureSkipVerify, opts.UpstreamURL, opts.TLSCertFile, opts.TLSKeyFile, opts.OpenshiftCAs)
	if err != nil {
		return nil, err
	}
	return &SearchGuardClient{esClient}, nil
}

func (sg *SearchGuardClient) FetchRoles() (*searchguard.Roles, error) {
	resp, err := sg.esClient.Get("/sg/roles")
	if err != nil {
		return nil, err
	}
	var data interface{}
	json.Unmarshal([]byte(resp), &data)
	encodedRoles, err := rolesLookup.Lookup(data)
	if err != nil {
		return nil, err
	}
	sRoles, err := base64.StdEncoding.DecodeString(encodedRoles.(string))
	if err != nil {
		return nil, err
	}
	roles := &searchguard.Roles{}
	err = roles.FromJson(string(sRoles))
	if err != nil {
		return nil, err
	}
	return roles, nil
}
func (sg *SearchGuardClient) FetchRolesMapping() (*searchguard.RolesMapping, error) {
	resp, err := sg.esClient.Get("/sg/rolesmapping")
	if err != nil {
		return nil, err
	}
	var data interface{}
	json.Unmarshal([]byte(resp), &data)
	encoded, err := rolesLookup.Lookup(data)
	if err != nil {
		return nil, err
	}
	sRolesMapping, err := base64.StdEncoding.DecodeString(encoded.(string))
	if err != nil {
		return nil, err
	}
	rolesmapping := &searchguard.RolesMapping{}
	err = rolesmapping.FromJson(string(sRolesMapping))
	if err != nil {
		return nil, err
	}
	return rolesmapping, nil
}

func (sg *SearchGuardClient) FlushRolesMapping(rolesMapping *searchguard.RolesMapping) error {
	json, err := rolesMapping.ToJson()
	if err != nil {
		return err
	}
	var out []byte
	base64.StdEncoding.Encode(out, []byte(json))
	doc := map[string]string{"role": string(out)}
	sDoc, err := searchguard.ToJson(doc)
	if err != nil {
		return err
	}
	sg.esClient.Put("sg/rolesmapping", sDoc)
	return nil
}

func (sg *SearchGuardClient) FlushRoles(roles *searchguard.Roles) error {
	json, err := roles.ToJson()
	if err != nil {
		return err
	}
	var out []byte
	base64.StdEncoding.Encode(out, []byte(json))
	doc := map[string]string{"role": string(out)}
	sDoc, err := searchguard.ToJson(doc)
	if err != nil {
		return err
	}
	sg.esClient.Put("sg/roles", sDoc)
	return nil
}
