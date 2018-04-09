GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: check

fmt: ## Run gofmt across all go files
	gofmt -w $(GOFMT_FILES)

dev:: ## Build the CLI
	mkdir -p ./bin
	govvv build -o bin/tsg-cli ./cmd/tsg-cli

release: default ## Making release build of the API
	@goreleaser --rm-dist --release-notes=CHANGELOG.md

tools:: ## Download and install all dev/code tools
	@echo "==> Installing dev/build tools"
	go get -u github.com/ahmetb/govvv
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/goreleaser/goreleaser
	gometalinter --install

check:: ## Run the code through metalinter
	gometalinter \
			--skip=examples \
			--deadline 10m \
			--vendor \
			--sort="path" \
			--aggregate \
			--enable-gc \
			--disable-all \
			--enable goimports \
			--enable misspell \
			--enable vet \
			--enable deadcode \
			--enable varcheck \
			--enable ineffassign \
			--enable gofmt \
			./...

.PHONY: help
help:: ## Display this help message
	@echo "GNU make(1) targets:"
	@grep -E '^[a-zA-Z_.-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
