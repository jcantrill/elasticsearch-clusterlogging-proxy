package clients

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"os"

	"github.com/bitly/go-simplejson"

	"github.com/openshift/elasticsearch-clusterlogging-proxy/util"
)

type OpenShiftClient struct {
	httpClient *http.Client
	token      string
}

func (c *OpenShiftClient) Get(path string) (*simplejson.Json, error) {
	req, err := http.NewRequest("GET", getKubeAPIURLWithPath(path).String(), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return nil, err
	}
	return request(c.httpClient, req)
}

func request(client *http.Client, req *http.Request) (*simplejson.Json, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%s %s %s", req.Method, req.URL, err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("%d %s %s %s", resp.StatusCode, req.Method, req.URL, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("got %d %s", resp.StatusCode, body)
	}
	data, err := simplejson.NewJson(body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// copy of same function in provider.go
func getKubeAPIURLWithPath(path string) *neturl.URL {
	ret := &neturl.URL{
		Scheme: "https",
		Host:   "kubernetes.default.svc",
		Path:   path,
	}

	if host := os.Getenv("KUBERNETES_SERVICE_HOST"); len(host) > 0 {
		ret.Host = host
	}

	return ret
}

// NewOpenShiftClient returns a client for connecting to the master.
func NewOpenShiftClient(paths []string, token string) (*OpenShiftClient, error) {
	if token == "" {
		return nil, fmt.Errorf("Unable to make requests to api server using a blank user token")
	}
	//defaults
	capaths := []string{"/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"}
	systemRoots := true
	if len(paths) != 0 {
		capaths = paths
		systemRoots = false
	}
	pool, err := util.GetCertPool(capaths, systemRoots)
	if err != nil {
		return nil, err
	}

	return &OpenShiftClient{
		&http.Client{
			Jar: http.DefaultClient.Jar,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
			},
		},
		token,
	}, nil
}
