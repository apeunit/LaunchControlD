GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
GIT_DESCR = $(shell git describe --tags --always)
APP=lctrld
# build output folder
OUTPUTFOLDER = dist
# docker image
DOCKER_REGISTRY = noandrea
DOCKER_IMAGE = LaunchControlD
DOCKER_TAG = $(GIT_DESCR)
# build paramters
OS = linux
ARCH = amd64
# K8S
K8S_NAMESPACE = geo
K8S_DEPLOYMENT = LaunchControlD
# SSH
SSH_HOST = evtvz-one
SSH_PATH = /usr/local/bin
SSH_SERVICE = lctrld

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs


default: build

build: build-dist

build-dist: $(GOFILES)
	@echo build binary to $(OUTPUTFOLDER)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static" -X main.Version=$(GIT_DESCR)' -o $(OUTPUTFOLDER)/$(APP) .
	@echo copy resources
	cp -r README.md LICENSE $(OUTPUTFOLDER)
	@echo done

build-zip: build
	@echo build zip release
	zip -rmT $(APP)-$(GIT_DESCR).zip $(OUTPUTFOLDER)
	sha1sum $(APP)-$(GIT_DESCR).zip
	@echo done

install: build
	cp dist/lctrld $(GOPATH)/bin
	@echo done

test: test-all

test-all:
	@go test $(GOPACKAGES) -v -race -coverprofile=cover.out -covermode=atomic

bench: bench-all

bench-all:
	@go test -bench -v $(GOPACKAGES)

lint: lint-all

lint-all:
	golint -set_exit_status $(GOPACKAGES)
	staticcheck $(GOPACKAGES)

clean:
	@echo remove $(OUTPUTFOLDER) folder
	rm -rf $(OUTPUTFOLDER)
	@echo done

doc:
	@echo generate the documentation 
	swag init -g pkg/server/web.go -o api
	@echo done

docker: docker-build

docker-build:
	@echo copy resources
	docker build --build-arg DOCKER_TAG='$(GIT_DESCR)' -t $(DOCKER_IMAGE)  .
	@echo done

docker-push:
	@echo push image
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo done

docker-run: 
	docker run -p 2004:2004 $(DOCKER_IMAGE):latest

debug-start:
	@go run main.go start

deploy-ssh: clean build-dist
	@echo deploy to $(SSH_HOST)
	scp $(OUTPUTFOLDER)/$(APP) $(SSH_HOST):$(SSH_PATH)/$(APP).upl
	ssh -t $(SSH_HOST) "mv $(SSH_PATH)/$(APP).upl $(SSH_PATH)/$(APP); systemctl restart $(SSH_SERVICE)"
	@echo deploy complete

deploy-k8s:
	@echo deploy k8s
	kubectl -n $(K8S_NAMESPACE) set image deployment/$(K8S_DEPLOYMENT) $(DOCKER_IMAGE)=$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo done

rollback-k8s:
	@echo deploy k8s
	kubectl -n $(K8S_NAMESPACE) rollout undo deployment/$(K8S_DEPLOYMENT)
	@echo done

changelog:
	git-chglog --output CHANGELOG.md

git-release:
	@echo making release
	git tag $(GIT_DESCR)
	git-chglog --output CHANGELOG.md
	git tag $(GIT_DESCR) --delete
	git add CHANGELOG.md && git commit -m "$(GIT_DESCR)" -m "Changelog: https://github.com/noandrea/$(APP)/blob/master/CHANGELOG.md"
	git tag -a "$(GIT_DESCR)" -m "Changelog: https://github.com/apeunit/$(APP)/blob/master/CHANGELOG.md"
	@echo release complete


_release-patch:
	$(eval GIT_DESCR = $(shell git describe --tags | awk -F '("|")' '{ print($$1)}' | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
release-patch: _release-patch git-release

_release-minor:
	$(eval GIT_DESCR = $(shell git describe --tags | awk -F '("|")' '{ print($$1)}' | awk -F. '{$$(NF-1) = $$(NF-1) + 1;} 1' | sed 's/ /./g' | awk -F. '{$$(NF) = 0;} 1' | sed 's/ /./g'))
release-minor: _release-minor git-release

_release-major:
	$(eval GIT_DESCR = $(shell git describe --tags | awk -F '("|")' '{ print($$1)}' | awk -F. '{$$(NF-2) = $$(NF-2) + 1;} 1' | sed 's/ /./g' | awk -F. '{$$(NF-1) = 0;} 1' | sed 's/ /./g' | awk -F. '{$$(NF) = 0;} 1' | sed 's/ /./g' ))
release-major: _release-major git-release 
