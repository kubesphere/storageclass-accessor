VERSION = v0.1.0
GIT_COMMIT=$(shell git rev-parse HEAD | head -c 7)

IMG ?= stoneshiyunify/storage-class-accessor:${GIT_COMMIT}
IMGLATEST ?= stoneshiyunify/storage-class-accessor:latest

.PHONY: build
build:
	go mod tidy && go mod verify && go build -o bin/manager main.go

.PHONY: docker-build
docker-build: #test ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}
	docker tag ${IMG} ${IMGLATEST}
	docker push ${IMGLATEST}