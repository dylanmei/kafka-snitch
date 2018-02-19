TARGET ?= $(shell basename `git rev-parse --show-toplevel`)
TESTS ?= $(shell go list ./... | grep -v /vendor/)

check-var = $(if $(strip $($1)),,$(error var for "$1" is empty))

default: test build

test:
	go test -v -cover -run=$(RUN) $(TESTS)

build: clean
	@go build -v -o bin/$(TARGET) .

release: test clean
	$(call check-var,REV)

	@CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
		-a -tags netgo \
		-a -installsuffix cgo \
		-ldflags "-X main.Revision=$(REV)" \
		-o bin/release/$(TARGET) .

docker/build: release
	@docker build -t kafka-snitch .

docker/push: docker/build
	$(call check-var,TAG)

	@docker tag kafka-snitch dylanmei/kafka-snitch:$(TAG)
	@docker push dylanmei/kafka-snitch:$(TAG)

clean:
	@rm -rf bin/
