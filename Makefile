GOFLAGS :=
IMAGE_REPOSITORY_NAME ?= openshift
BIN_NAME=elasticsearch-clusterlogging-proxy

build:
	go build $(GOFLAGS) .
.PHONY: build

images:
	imagebuilder -f Dockerfile -t $(IMAGE_REPOSITORY_NAME)/$(BIN_NAME) .
.PHONY: images

clean:
	$(RM) ./$(BIN_NAME)
.PHONY: clean
