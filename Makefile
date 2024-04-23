test:
	@echo "Running Golang unit tests"
	@go test -v -short ./... -skip Example

lint:
	@echo "Running Golang Linter"
	@golangci-lint run
	@echo "Running Markdown Linter"
	@markdownlint *.md
