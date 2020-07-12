########################################
# stage 1: build web app
########################################
FROM node:10.16.0-alpine as web-builder

WORKDIR /app

COPY client/package.json .
COPY client/yarn.lock .
RUN yarn

COPY ./client .
RUN rm -fr build && yarn build

########################################
# stage 2: build go
########################################
FROM golang:1.14 as go-builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN rm -fr ./dist && mkdir ./dist
RUN cp ./config/prod.yml ./dist/prod.yml
RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -o ./dist/pushaas main.go

########################################
# stage 3: run
########################################
FROM alpine:latest

WORKDIR /app

EXPOSE 8080

ENV PUSHAPI_ENV=prod

COPY --from=go-builder /app/dist/pushaas ./pushaas
COPY --from=go-builder /app/config/prod.yml ./config/prod.yml
COPY --from=web-builder /app/build ./client/build

# thanks https://github.com/aws/aws-sdk-go/issues/2322#issuecomment-443502850
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ENTRYPOINT ["/app/pushaas"]
