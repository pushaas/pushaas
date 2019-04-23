.PHONY: build \
	run

########################################
# app
########################################
build:
	@go build -o dist/pushaas pushaas/main.go

run:
	@# TODO: load credentials and profile from environment variables
	@AWS_PROFILE=pushaas AWS_SDK_LOAD_CONFIG=true go run pushaas/main.go

watch:
	@realize start --run --no-config

########################################
# docker
########################################
# docker-build-prod:

docker-build-dev:
	@docker build \
		-f Dockerfile-dev \
		-t pushaas:latest \
		.

docker-run: docker-build-dev
	@echo "done"
	@docker run -it -p 9000:9000 pushaas:latest

# docker-push:

########################################
# services
########################################
# services-up:
