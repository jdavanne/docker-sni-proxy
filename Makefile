VERSION := 0.0.2
NAME := docker-sni-proxy
DATE := $(shell date +'%Y-%M-%d_%H:%M:%S')
BUILD := $(shell git rev-parse HEAD | cut -c1-8)
LDFLAGS :=-ldflags "-s -w -X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -X=main.Date=$(DATE)"
IMAGE := $(NAME)
REGISTRY := davinci1976
PUBLISH := $(REGISTRY)/$(IMAGE)

.PHONY: docker all deps

all: build

build:
	(cd src ; go build -o ../$(NAME) $(LDFLAGS))

docker-test:
	docker-compose -f docker-compose.test.yml down
	docker-compose -f docker-compose.test.yml build
	docker-compose -f docker-compose.test.yml up --abort-on-container-exit || (docker-compose -f docker-compose.test.yml down ; exit 1)
	#docker-compose -f docker-compose.test.yml run sut  || (docker-compose -f docker-compose.test.yml logs -t | sort -k 3 ; docker-compose -f docker-compose.test.yml down ; exit 1)
	docker-compose -f docker-compose.test.yml down

deps:
	go list -f '{{range .TestImports}}{{.}} {{end}} {{range .Imports}}{{.}} {{end}}' ./... | tr ' ' '\n' | grep -e "^[^/_\.][^/]*\.[^/]*/" |sort -u >.deps

deps-install:
	go get -v $$(cat .deps)
	#for dep in $$(cat .deps); do echo "installing '$$dep'... "; go get -v $$dep; done

deps-install-force: deps
	go get -u -v $$(cat .deps)
	#for dep in $$(cat .deps); do echo "installing '$$dep'... "; go get -u -v $$dep; done

docker-run:
	docker-compose up

docker:
	docker build -t $(IMAGE) .

docker-publish-all: docker-publish docker-publish-version

docker-publish-version:
	docker tag $(IMAGE) $(PUBLISH):$(VERSION)
	docker push $(PUBLISH):$(VERSION)

docker-publish: docker
	docker tag $(IMAGE) $(PUBLISH):latest
	docker push $(PUBLISH):latest
