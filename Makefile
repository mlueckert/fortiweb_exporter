VERSION  ?= $(shell git describe --tags --always --dirty)
GIT_HASH ?= $(shell git rev-parse --short HEAD)
# -trimpath and -s -w: reproducible builds without local paths or debug symbols.
LDFLAGS   = -trimpath -ldflags "-s -w -X main.Version=$(VERSION) -X main.GitHash=$(GIT_HASH)"
BINARY    = fortiweb-exporter
GOVULNCHECK = golang.org/x/vuln/cmd/govulncheck@latest
BETTERLEAKS = github.com/betterleaks/betterleaks@v1.8.1
TARGET    = target

.PHONY: build
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(TARGET)/$(BINARY) .

# Cross-compiled release assets, uploaded by semantic-release (see .releaserc.yml).
.PHONY: build-release
build-release:
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(TARGET)/$(BINARY).linux.amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o $(TARGET)/$(BINARY).linux.arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(TARGET)/$(BINARY).windows.amd64.exe .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(TARGET)/$(BINARY).darwin.amd64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(TARGET)/$(BINARY).darwin.arm64 .

.PHONY: clean
clean:
	rm -rf $(TARGET)

.PHONY: fmt-check
fmt-check:
	@FILES="$$(gofmt -l .)"; \
	if [ -n "$$FILES" ]; then \
		printf "Files not properly formatted, run 'gofmt -w .':\n%s\n" "$$FILES"; \
		exit 1; \
	fi

.PHONY: vet
vet:
	go vet ./...

# Reports known vulnerabilities in dependencies and the Go standard library
# that are reachable from this code. Needs network access to vuln.go.dev.
.PHONY: vulncheck
vulncheck:
	go run $(GOVULNCHECK) ./...

# Scans the full git history for leaked secrets.
.PHONY: secrets-check
secrets-check:
	go run $(BETTERLEAKS) git . --redact --verbose

# Installs the git hooks from .githooks (betterleaks pre-commit secret scan).
.PHONY: hooks
hooks:
	git config core.hooksPath .githooks

.PHONY: test
test: fmt-check vet
	go test -race ./...
