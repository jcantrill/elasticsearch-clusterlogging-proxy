package clients

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/oliveagle/jsonpath"
	ext "github.com/openshift/elasticsearch-clusterlogging-proxy/extensions"
	"github.com/openshift/elasticsearch-clusterlogging-proxy/extensions/clusterlogging/searchguard"
	log "github.com/sirupsen/logrus"
)

var (
	rolesLookup, _        = jsonpath.Compile("$._source.roles")
	rolesMappingLookup, _ = jsonpath.Compile("$._source.rolesmapping")
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

func decodeACLDocument(resp string, matcher *jsonpath.Compiled) (string, error) {
	var data interface{}
	json.Unmarshal([]byte(resp), &data)
	encoded, err := matcher.Lookup(data)
	if err != nil {
		return "", err
	}
	unencoded, err := base64.StdEncoding.DecodeString(encoded.(string))
	if err != nil {
		return "", err
	}
	return string(unencoded), nil
}

func (sg *SearchGuardClient) FetchRoles() (*searchguard.Roles, error) {
	log.Debug("Fetching SG roles...")
	resp, err := sg.esClient.Get("/.searchguard/sg/roles")
	if err != nil {
		return nil, err
	}
	sRoles, err := decodeACLDocument(resp, rolesLookup)
	if err != nil {
		return nil, err
	}
	roles := &searchguard.Roles{}
	err = roles.FromJson(sRoles)
	if err != nil {
		return nil, err
	}
	log.Debugf("Roles: %s", sRoles)
	return roles, nil
}

func (sg *SearchGuardClient) FetchRolesMapping() (*searchguard.RolesMapping, error) {
	log.Debug("Fetching SG rolesmapping...")
	resp, err := sg.esClient.Get("/.searchguard/sg/rolesmapping")
	if err != nil {
		return nil, err
	}
	sRolesMapping, err := decodeACLDocument(resp, rolesMappingLookup)
	if err != nil {
		return nil, err
	}
	rolesmapping := &searchguard.RolesMapping{}
	err = rolesmapping.FromJson(sRolesMapping)
	if err != nil {
		return nil, err
	}
	log.Debugf("Rolesmapping: %s", sRolesMapping)
	return rolesmapping, nil
}

func encodeACLDocument(doc searchguard.Serializable) (string, error) {
	log.Tracef("Encoding %s ACL Document...", doc.Type())
	json, err := doc.ToJson()
	if err != nil {
		return "", err
	}
	log.Tracef("Trying to encode: %s", json)
	encodedData := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, encodedData)
	defer encoder.Close()
	encoder.Write([]byte(json))
	updated := map[string]string{doc.Type(): encodedData.String()}
	return searchguard.ToJson(updated)
}

func (sg *SearchGuardClient) FlushACL(doc searchguard.Serializable) error {
	log.Tracef("Flushing SG %s: %+v", doc.Type(), doc)
	sDoc, err := encodeACLDocument(doc)
	if err != nil {
		return err
	}
	if _, err = sg.esClient.Put(fmt.Sprintf("/.searchguard/sg/%s", doc.Type()), sDoc); err != nil {
		return err
	}
	return nil
}
