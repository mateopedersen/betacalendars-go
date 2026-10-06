# Contributing

Contributions should preserve deterministic behavior, keep the module dependency-free unless a dependency is necessary, and include tests for externally visible behavior.

Before opening a pull request, run:

```sh
gofmt -w $(find . -name '*.go' -type f)
go mod tidy
go vet ./...
go test ./...
go test -race ./...
```

Keep exported identifiers documented and update README or changelog entries when public behavior changes. Do not rewrite or move published version tags.
