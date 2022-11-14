.PHONY: vendor
vendor: vendor/modules.txt

vendor/modules.txt: go.mod
	go mod vendor

.PHONY: build
build: clair-action

clair-action: vendor
	go build -o clair-action ./cmd/...
