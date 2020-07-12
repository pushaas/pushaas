TAG := latest
CONTAINER := pushaas
IMAGE := pushaas/$(CONTAINER)
IMAGE_TAGGED := $(IMAGE):$(TAG)
NETWORK := pushaas_default
PORT_CONTAINER := 9000
PORT_HOST := 9000

CONTAINER_DEV := $(CONTAINER)-dev
IMAGE_DEV := pushaas/$(CONTAINER_DEV)
IMAGE_TAGGED_DEV := $(IMAGE_DEV):$(TAG)

REDIS_TAG := latest
CONTAINER_REDIS := pushaas-redis
IMAGE_REDIS := pushaas/$(CONTAINER_REDIS)
IMAGE_TAGGED_REDIS := $(IMAGE_REDIS):$(REDIS_TAG)

########################################
# app
########################################
.PHONY: setup
setup:
	@go get github.com/onsi/ginkgo/ginkgo
	@go get github.com/onsi/gomega/...
	@go get github.com/matryer/moq

.PHONY: clean
clean:
	@rm -fr ./dist

.PHONY: build
build: clean
	@go build -o ./dist/pushaas main.go

.PHONY: run
run:
	@AWS_PROFILE=pushaas AWS_SDK_LOAD_CONFIG=true go run main.go

.PHONY: kill
kill:
	@-killall push-api

.PHONY: setup-client
setup-client:
	@cd client && yarn

.PHONY: build-client
build-client:
	@cd client && yarn build

########################################
# test
########################################
.PHONY: test
test:
	@GIN_MODE=release ginkgo -r -keepGoing -focus=$(FOCUS)

.PHONY: test-watch
test-watch:
	@GIN_MODE=release ginkgo watch -r -depth=0 -focus="${FOCUS}"

.PHONY: test-coverage
test-coverage:
	@go clean
	@rm -f ./coverage.out
	-@GIN_MODE=release ginkgo -r -keepGoing -cover -coverprofile=coverage.out -outputdir=.
	@for i in */**/coverage.out; do rm -f "$i"; done
	@grep -v "mode" coverage.out > temp_coverage.out && mv temp_coverage.out coverage.out
	@echo "mode: set" | cat - coverage.out> temp_coverage.out && mv temp_coverage.out coverage.out
	@go tool cover -html=coverage.out

.PHONY: test-generate-mocks
test-generate-mocks: test-generate-pushaas-mocks test-generate-library-mocks

.PHONY: test-generate-pushaas-mocks
test-generate-pushaas-mocks:
	@moq -out pushaas/mocks/bind_service.go -pkg mocks pushaas/services BindService
	@moq -out pushaas/mocks/instance_service.go -pkg mocks pushaas/services InstanceService
	@moq -out pushaas/mocks/plan_service.go -pkg mocks pushaas/services PlanService
	@moq -out pushaas/mocks/provision_service.go -pkg mocks pushaas/services ProvisionService

.PHONY: test-generate-library-mocks
test-generate-library-mocks:
	@moq -out pushaas/mocks/redis_client.go -pkg mocks ${GOPATH}/src/github.com/go-redis/redis UniversalClient

########################################
# docker
########################################

# dev
.PHONY: docker-clean-dev
docker-clean-dev:
	@-docker rm -f $(CONTAINER_DEV)

.PHONY: docker-build-dev
docker-build-dev:
	@docker build \
		-f Dockerfile-dev \
		-t $(IMAGE_TAGGED_DEV) \
		.

.PHONY: docker-run-dev
	docker-run-dev: docker-clean-dev
	@docker run \
		-it \
		--name=$(CONTAINER_DEV) \
		--network=$(NETWORK) \
		-p $(PORT_HOST):$(PORT_CONTAINER) \
		$(IMAGE_TAGGED_DEV)

.PHONY: docker-build-and-run-dev
docker-build-and-run-dev: docker-build-dev docker-run-dev

# prod
.PHONY: docker-clean
docker-clean:
	@-docker rm -f $(CONTAINER)

.PHONY: docker-build
docker-build:
	@docker build \
		-f Dockerfile \
		-t $(IMAGE_TAGGED) \
		.

.PHONY: docker-run
docker-run: docker-clean
	@docker run \
		-it \
		--name=$(CONTAINER) \
		--network=$(NETWORK) \
		-p $(PORT_HOST):$(PORT_CONTAINER) \
		$(IMAGE_TAGGED)

.PHONY: docker-build-and-run
docker-build-and-run: docker-build docker-run

.PHONY: docker-push
docker-push: docker-build
	@docker push \
		$(IMAGE_TAGGED)

# redis
.PHONY: docker-clean-redis
docker-clean-redis:
	@-docker rm -f $(CONTAINER_REDIS)

.PHONY: docker-build-redis
docker-build-redis:
	@docker build \
		-f Dockerfile-redis \
		-t $(IMAGE_TAGGED_REDIS) \
		.

.PHONY: docker-push-redis
docker-push-redis: docker-build-redis
	@docker push \
		$(IMAGE_TAGGED_REDIS)

########################################
# services
########################################
.PHONY: services-up
services-up:
	@docker-compose up -d --remove-orphans

.PHONY: services-down
services-down:
	@docker-compose down --remove-orphans
