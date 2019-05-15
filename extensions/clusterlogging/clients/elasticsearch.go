package clients

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
	resp, err := es.client.Get(url(es.serverURL, path))
	if err != nil {
		return "", err
	}
	return readBody(resp)
}

func readBody(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return string(body), nil
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
