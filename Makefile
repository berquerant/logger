.PHONY: test
test:
	go test -cover ./...

.PHONY: go-regenerate
go-regenerate: clean-go-generate go-generate

.PHONY: clean-go-generate
clean-go-generate:
	find $(ROOT) -name "*_generated.go" -type f | xargs rm -f

.PHONY: go-generate
go-generate:
	go generate ./...
