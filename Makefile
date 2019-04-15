.PHONY: build \
	run

build:
	@go build -o dist/pushaas main.go

run:
	@go run main.go

watch:
	@realize start --run --no-config
