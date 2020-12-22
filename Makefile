SLUG = $(shell head -1 go.mod | cut -d/ -f2-)
NAME = mp707

CGO_ENABLED = 1
GO111MODULE = on

export SLUG
export NAME
export CGO_ENABLED
export GO111MODULE

LDFLAGS += -w -s

.PHONY: all
all: build

.PHONY: check
check: vet test

.PHONY: vet
vet:
	go vet -v ./...

.PHONY: test
test:
	go test -v -count 1 ./...

.PHONY: clean
clean:
	rm -vf cmd/$(NAME)/$(NAME)
	rm -rvf release

.PHONY: build
build: cmd/$(NAME)/$(NAME)

.PHONY: static
static: LDFLAGS += -linkmode external -extldflags '-static $(shell pkg-config --libs libudev)'
static: build

cmd/$(NAME)/$(NAME):
	cd cmd/$(NAME); go build -v -ldflags "$(LDFLAGS)"

release: release/$(NAME).linux.386 release/$(NAME).linux.amd64

release/$(NAME).linux.%:
	GOARCH=$* OUTPUT=$@ scripts/release.sh
