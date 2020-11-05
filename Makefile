.PHONY: release
release:
	goreleaser --rm-dist

.PHONY: pre-release
pre-release:
	goreleaser --snapshot --skip-publish --rm-dist

.PHONY: test
test: test-unit test-integration

.PHONY: test-integration
test-integration:
	go test -tags integration .

.PHONY: test-unit
test-unit:
	go test -race .
