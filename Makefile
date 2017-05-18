VERSION := 0.0.1
NAME := docker-sni-proxy
DATE := $(shell date +'%Y-%M-%d_%H:%M:%S')
BUILD := $(shell git rev-parse HEAD | cut -c1-8)
LDFLAGS :=-ldflags "-s -w -X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -X=main.Date=$(DATE)"
IMAGE := jdavanne/$(NAME)
.PHONY: docker all

all: build

build:
	(cd src ; go build -o ../$(NAME) $(LDFLAGS))

deps:
	go list -f '{{range .TestImports}}{{.}} {{end}} {{range .Imports}}{{.}} {{end}}' ./... | sed 's/ /\n/g' | grep -e "^[^/_\.][^/]*\.[^/]*/" |sort -u >.deps

deps-install:
	go get -v $$(cat .deps)
	#for dep in $$(cat .deps); do echo "installing '$$dep'... "; go get -v $$dep; done

deps-install-force: deps
	go get -u -v $$(cat .deps)
	#for dep in $$(cat .deps); do echo "installing '$$dep'... "; go get -u -v $$dep; done

docker:
	docker build -t $(IMAGE) .

test:
	./test.sh
