.PHONY: l
l: ## Lint Go source files
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && golangci-lint run

.PHONY: t
t: ## Run unit tests
	go test -race -count=1 ./...

.PHONY: f
f: ## Format code
	go fmt ./...

.PHONY: c
c: ## Measure code coverage
	go test -race -covermode=atomic ./... -coverprofile=cover.out && \
	go tool cover -func cover.out \
		| grep -v '100.0%' \
		|| true

	rm cover.out
