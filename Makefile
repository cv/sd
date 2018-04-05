BINARY = sd
VERSION = 0.0.3
COMMIT = $(shell git rev-parse --short HEAD)

PLATFORMS := windows linux darwin
os = $(word 1, $@)

.PHONY: build
build: windows linux darwin

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	@GOOS=$(os) GOARCH=amd64 go build -ldflags='-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)' -v -o build/$(BINARY)-$(VERSION)-$(os)-amd64

clean:
	@rm -rf build/sd-*
