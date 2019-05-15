package clients

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/oliveagle/jsonpath"
	log "github.com/sirupsen/logrus"
)

var (
	errorLookup, _ = jsonpath.Compile("$.error")
)

//ElasticsearchClient is an admin client to query a local instance of Elasticsearch
type ElasticsearchClient struct {
	serverURL string
	client    *http.Client
}

//NewElasticsearchClient is the initializer to create an instance of ES client
func NewElasticsearchClient(skipVerify bool, serverURL, adminCert, adminKey string, adminCA []string) (*ElasticsearchClient, error) {
	caCertPool := x509.NewCertPool()
	for _, ca := range adminCA {
		caCert, err := ioutil.ReadFile(ca)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool.AppendCertsFromPEM(caCert)
	}

	cert, err := tls.LoadX509KeyPair(adminCert, adminKey)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: skipVerify,
			},
		},
	}
	return &ElasticsearchClient{serverURL, client}, nil
}

func url(elasticsearchURL, path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return elasticsearchURL + path
}

//Get the content at the path
func (es *ElasticsearchClient) Get(path string) (string, error) {
	url := url(es.serverURL, path)
	log.Tracef("Get: %v", url)
	resp, err := es.client.Get(url)
	log.Tracef("Response code: %v", resp.StatusCode)
	if err != nil {
		return "", err
	}
	body, err := readBody(resp)
	if err != nil {
		return "", err
	}
	return body, nil
}

func readBody(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result := string(body)
	log.Tracef("Response body: %s", result)
	if resp.StatusCode != 200 {
		log.Trace("Additionally inspecting result of non 200 response...")
		var data interface{}
		json.Unmarshal([]byte(body), &data)
		errorBody, err := errorLookup.Lookup(data)
		log.Tracef("errBody: %v", errorBody)
		if err == nil {
			return errorBody.(string), nil
		}
		log.Tracef("Error trying to decode json response %v", err)
	}
	return result, nil
}

//Put submits a PUT request to ES assuming the given body is of type 'application/json'
func (es *ElasticsearchClient) Put(path string, body string) (string, error) {
	request, err := http.NewRequest("PUT", url(es.serverURL, path), strings.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	resp, err = es.client.Do(request)
	if err != nil {
		return "", err
	}
	return readBody(resp)
}
