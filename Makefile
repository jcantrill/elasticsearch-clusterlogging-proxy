GOFLAGS :=
IMAGE_REPOSITORY_NAME ?= openshift
BIN_NAME=elasticsearch-clusterlogging-proxy

#inputs to 'run' which may need to change
TLS_CERTS_BASEDIR=_output
CLIENT_SECRET?=SzVEeEQwYmRFcVRpb3VaWVpFUmdKbjN3bnZweWxrR3FRU1RWY01BSWNTdDRPRk9wYkdaMjB4cWN6ODRhMElFUg==
COOKIE_SECRET?=3bM3IXYGSivKBWW+xE1uQg==

PKGS=$(shell go list ./... | grep -v -E '/vendor/')
TEST_OPTIONS?=

build:
	go build $(GOFLAGS) .
.PHONY: build

images:
	imagebuilder -f Dockerfile -t $(IMAGE_REPOSITORY_NAME)/$(BIN_NAME) .
.PHONY: images

clean:
	$(RM) ./$(BIN_NAME)
	rm -rf $(TLS_CERTS_BASEDIR)
.PHONY: clean

test:
	@go test $(TEST_OPTIONS) $(PKGS)
.PHONY: test

prep-for-run:
	mkdir -p ${TLS_CERTS_BASEDIR}||:  && \
	for n in "ca" "cert" "key" ; do \
		oc get secret elasticsearch -o jsonpath={.data.admin-$$n} | base64 -d > _output/admin-$$n ; \
	done && \
	oc get pod -l component=elasticsearch -o jsonpath={.items[0].metadata.name} > _output/espod && \
	oc exec -c elasticsearch $$(cat _output/espod) -- cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt > _output/ca.crt && \
	oc serviceaccounts get-token elasticsearch > _output/sa-token && \
	echo openshift-logging > _output/namespace && \
	mkdir -p /var/run/secrets/kubernetes.io/serviceaccount/||:  && \
	sudo ln -sf $${PWD}/_output/ca.crt /var/run/secrets/kubernetes.io/serviceaccount/ca.crt && \
	sudo ln -sf $${PWD}/_output/sa-token /var/run/secrets/kubernetes.io/serviceaccount/token && \
	sudo ln -sf $${PWD}/_output/namespace /var/run/secrets/kubernetes.io/serviceaccount/namespace
	
.PHONY: prep-for-run

run:
	./$(BIN_NAME) --https-address=':60000' \
        --provider=openshift \
        --upstream=https://127.0.0.1:9200 \
        --tls-cert=$(TLS_CERTS_BASEDIR)/admin-cert \
        --tls-key=$(TLS_CERTS_BASEDIR)/admin-key \
        --upstream-ca=$(TLS_CERTS_BASEDIR)/admin-ca \
        --openshift-service-account=elasticsearch \
		--openshift-delegate-urls='{"/": {"resource": "namespaces", "verb": "get"}}' \
        --cookie-secret=$(COOKIE_SECRET) \
        --pass-user-bearer-token \
        --skip-provider-button \
		--ssl-insecure-skip-verify
.PHONY: run
