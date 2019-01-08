.DEFAULT_GOAL := all

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)

COMMON_LDFLAGS = -X main.GitCommit='$(GIT_COMMIT)$(GIT_DIRTY)'
COMMON_ARGS = -ldflags "$(COMMON_LDFLAGS)"

DEPS = $(wildcard *.go) $(wildcard cmd/xenstore/*.go)

OUTPUTS = \
	xenstore-linux-amd64 \
	xenstore-linux-amd64-static \
	xenstore-windows-amd64.exe

.PHONY: all
all: $(OUTPUTS)

.PHONY: clean
clean:
	rm -rf $(OUTPUTS)

.PHONY: update
update:
	dep ensure -update
	dep prune

xenstore-linux-amd64: $(DEPS)
	env GOOS=linux GOARCH=amd64 go build $(COMMON_ARGS) -o $@ ./cmd/xenstore

xenstore-linux-amd64-static: $(DEPS)
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s $(COMMON_LDFLAGS)" -o $@ ./cmd/xenstore

xenstore-windows-amd64.exe: $(DEPS)
	env GOOS=windows GOARCH=amd64 go build $(COMMON_ARGS) -o $@ ./cmd/xenstore