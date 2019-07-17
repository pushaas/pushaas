TAG := latest
CONTAINER := pushaas
IMAGE := rafaeleyng/$(CONTAINER)
IMAGE_TAGGED := $(IMAGE):$(TAG)
NETWORK := pushaas_default
PORT_CONTAINER := 9000
PORT_HOST := 9000

CONTAINER_DEV := $(CONTAINER)-dev
IMAGE_DEV := rafaeleyng/$(CONTAINER_DEV)
IMAGE_TAGGED_DEV := $(IMAGE_DEV):$(TAG)

########################################
# app
########################################
.PHONY: setup
setup:
	@go get github.com/oxequa/realize

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

.PHONY: watch
watch: kill
	@AWS_PROFILE=pushaas AWS_SDK_LOAD_CONFIG=true realize start

.PHONY: build-client
build-client:
	@cd client && yarn build

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
.PHONY: docker-clean-prod
docker-clean-prod:
	@-docker rm -f $(CONTAINER)

.PHONY: docker-build-prod
docker-build-prod:
	@docker build \
		-f Dockerfile-prod \
		-t $(IMAGE_TAGGED) \
		.

.PHONY: docker-run-prod
docker-run-prod: docker-clean-prod
	@docker run \
		-it \
		--name=$(CONTAINER) \
		--network=$(NETWORK) \
		-p $(PORT_HOST):$(PORT_CONTAINER) \
		$(IMAGE_TAGGED)

.PHONY: docker-build-and-run-prod
docker-build-and-run-prod: docker-build-prod docker-run-prod

.PHONY: docker-push-prod
docker-push-prod: docker-build-prod
	@docker push \
		$(IMAGE_TAGGED)

########################################
# services
########################################
.PHONY: services-up
services-up:
	@docker-compose up -d --remove-orphans

.PHONY: services-down
services-down:
	@docker-compose down --remove-orphans

.PHONY: mongo-express
mongo-express:
	docker run -it --rm \
		--network $(NETWORK) \
		--name mongo-express \
		-p 8081:8081 \
		mongo-express
