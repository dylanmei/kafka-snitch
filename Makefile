TARGET   ?= $(shell basename `git rev-parse --show-toplevel`)
VERSION  ?= $(shell git describe --tags --always )
BRANCH   ?= $(shell git rev-parse --abbrev-ref HEAD)
REVISION ?= $(shell git rev-parse HEAD)
SHORTREV ?= $(shell git rev-parse --short HEAD)
LD_FLAGS ?= -s \
	-X github.com/dylanmei/kafka-snitch/version.Name=$(TARGET) \
	-X github.com/dylanmei/kafka-snitch/version.Revision=$(REVISION) \
	-X github.com/dylanmei/kafka-snitch/version.Branch=$(BRANCH) \
	-X github.com/dylanmei/kafka-snitch/version.Version=$(VERSION)

TESTS ?= $(shell go list ./... | grep -v /vendor/)

default: test build

test:
	go test -v -cover -run=$(RUN) $(TESTS)

build: clean
	@go build -v \
		-ldflags "$(LD_FLAGS)+local_changes" \
		-o bin/$(TARGET) .

release: test clean
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
		-a -tags netgo \
		-a -installsuffix cgo \
		-ldflags "$(LD_FLAGS)" \
		-o bin/release/$(TARGET) .

docker/build: release
	@docker build -t kafka-snitch .

docker/push: docker/build
	@docker tag kafka-snitch dylanmei/kafka-snitch:$(SHORTREV)
	@docker push dylanmei/kafka-snitch:$(SHORTREV)

clean:
	@rm -rf bin/
