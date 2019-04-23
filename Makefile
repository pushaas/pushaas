.PHONY: build \
	run

########################################
# app
########################################
build:
	@go build -o dist/pushaas main.go

run:
	@# TODO: load credentials and profile from environment variables
	@AWS_PROFILE=pushaas AWS_SDK_LOAD_CONFIG=true go run main.go

watch:
	@realize start --run --no-config

########################################
# docker
########################################
docker-build-prod:
	@docker build \
		-f Dockerfile-prod \
		-t rafaeleyng/pushaas:latest \
		.

docker-build-dev:
	@docker build \
		-f Dockerfile-dev \
		-t pushaas:latest \
		.

docker-run: docker-build-dev
	@docker run \
		-it \
		-p 9000:9000 \
		pushaas:latest

docker-push: docker-build-prod
	@docker push \
		rafaeleyng/pushaas

########################################
# services
########################################
services-up:
	@docker-compose up -d

services-down:
	@docker-compose down

mongo-express:
	docker run -it --rm \
		--network pushaas_default \
		--name mongo-express \
		-p 8081:8081 \
		mongo-express
