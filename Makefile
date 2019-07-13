.PHONY: build \
	run

########################################
# app
########################################
setup:
	@go get github.com/oxequa/realize

clean:
	@rm -fr ./dist

build: clean
#	@cp ./config/$(ENV).yml ./dist/config.yml
	@go build -o ./dist/pushaas main.go

run:
	@# TODO: in prod, load credentials and profile from environment variables
	@AWS_PROFILE=pushaas AWS_SDK_LOAD_CONFIG=true go run main.go

watch:
	@AWS_PROFILE=pushaas AWS_SDK_LOAD_CONFIG=true realize start --run --no-config

########################################
# docker
########################################

# dev
docker-build-dev:
	@docker build \
		-f Dockerfile-dev \
		-t pushaas-dev:latest \
		.

docker-run-dev:
	@docker run \
		-it \
		-p 9000:9000 \
		pushaas-dev:latest

docker-build-and-run-dev: docker-build-dev docker-run-dev

# prod
docker-build-prod:
	@docker build \
		-f Dockerfile-prod \
		-t rafaeleyng/pushaas:latest \
		.

docker-run-prod:
	@docker run \
		-it \
		-p 9000:9000 \
		rafaeleyng/pushaas:latest

docker-build-and-run-prod: docker-build-prod docker-run-prod

docker-push-prod: docker-build-prod
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
