ARG GO_VERSION=1.25.1-alpine
ARG ALPINE_VERSION=3.22

ARG BINARY_NAME=avito
ARG CONFIG_DIR=config
ARG WORKDIR_PATH=/app

FROM golang:${GO_VERSION} AS builder

ARG BINARY_NAME
ARG CONFIG_DIR

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${BINARY_NAME} ./cmd/

FROM alpine:${ALPINE_VERSION} AS app

ARG BINARY_NAME
ARG CONFIG_DIR
ARG WORKDIR_PATH

WORKDIR ${WORKDIR_PATH}
COPY --from=builder /build/${BINARY_NAME} .
RUN mkdir -p ${WORKDIR_PATH}/${CONFIG_DIR}
COPY --from=builder /build/${CONFIG_DIR}/ ./${CONFIG_DIR}/

EXPOSE 8080

# TODO: сделать путь с BINARY_NAME
CMD ["./avito"]
