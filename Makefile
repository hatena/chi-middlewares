.PHONY: help
help:
	@awk -F':.*##' '/^[-_a-zA-Z0-9]+:.*##/{printf"%-12s\t%s\n",$$1,$$2}' $(MAKEFILE_LIST) | sort

format: format-go format-markdown format-yaml ## format

format-go:
	go fmt -x ./...

format-markdown:
	git ls-files | grep -E '\.md$$' | xargs -t npx prettier --write

format-yaml:
	git ls-files | grep -E '\.ya?ml$$' | xargs -t npx prettier --write

lint: lint-markdown lint-yaml ## lint

lint-markdown:
	git ls-files | grep -E '\.md$$' | xargs -t npx prettier -c

lint-yaml:
	git ls-files | grep -E '\.ya?ml$$' | xargs -t npx prettier -c
	git ls-files | grep -E '\.ya?ml$$' | xargs -t yamllint

test: test-go ## test を実行する

test-go:
	go test -v -cover -race ./...
