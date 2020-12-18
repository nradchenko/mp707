NAME = mp707

CGO_ENABLED = 1
GO111MODULE = on

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

.PHONY: build
build: cmd/$(NAME)/$(NAME)

.PHONY: static
static: LDFLAGS += -linkmode external -extldflags '-static -ludev'
static: build

cmd/$(NAME)/$(NAME):
	cd cmd/$(NAME); go build -v -ldflags "$(LDFLAGS)"
