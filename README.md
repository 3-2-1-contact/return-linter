# Return Linter

A custom linter for golangci-lint that ensures `w.WriteHeader()` calls in HTTP middleware are immediately followed by `return` statements.

## Why This Linter?

In Go HTTP handlers, after calling `w.WriteHeader()`, you should typically return immediately to avoid unintended behavior. Continuing execution after writing the status code can lead to:

- Writing additional data to the response body unintentionally
- Confusion about the actual response being sent
- Potential security issues if error paths continue execution

## What It Checks

This linter specifically targets HTTP middleware functions with the following pattern:

```go
func MiddlewareName(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Handler code here
    })
}
```

Within these handlers, it checks that any call to `w.WriteHeader()` is immediately followed by a `return` statement.

## Examples

### ❌ Bad - Will Trigger Linter

```go
func BadMiddleware(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") == "" {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Unauthorized")) // This continues after WriteHeader!
        }
        handler.ServeHTTP(w, r)
    })
}
```

### ✅ Good - Correct Usage

```go
func GoodMiddleware(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") == "" {
            w.WriteHeader(http.StatusUnauthorized)
            return // Immediately returns after WriteHeader
        }
        handler.ServeHTTP(w, r)
    })
}
```

## Installation

### Using with golangci-lint (Plugin Method)

1. Build the plugin:
```bash
go build -buildmode=plugin -o returnlinter.so cmd/returnlinter/main.go
```

2. Add to your `.golangci.yml`:
```yaml
linters-settings:
  custom:
    returnlinter:
      path: ./returnlinter.so
      description: Checks that WriteHeader calls are followed by return statements
      original-url: github.com/jc/return-linter
```

3. Enable the linter:
```yaml
linters:
  enable:
    - returnlinter
```

### Standalone Usage

You can also run the linter directly using the `go vet` tool:

```bash
go vet -vettool=$(which returnlinter) ./...
```

## Building

```bash
# Install dependencies
go mod tidy

# Build the plugin
go build -buildmode=plugin -o returnlinter.so cmd/returnlinter/main.go

# Run tests
go test ./...
```

## Testing

The linter includes comprehensive test cases in `testdata/src/example/example.go`. Run tests with:

```bash
go test -v
```

## What Gets Checked

The linter checks for `WriteHeader()` calls in:
- Simple if/else blocks
- Switch/case statements
- Nested conditionals
- For/range loops

The linter **does not** check:
- Regular HTTP handler functions (not middleware)
- Functions that don't match the middleware pattern
- Functions that don't return `http.Handler`

## How It Works

The linter uses Go's AST (Abstract Syntax Tree) to:
1. Identify middleware functions matching the pattern `func(handler http.Handler) http.Handler`
2. Find `http.HandlerFunc` calls within those functions
3. Inspect the handler function body for `w.WriteHeader()` calls
4. Verify that the next statement after `WriteHeader()` is a `return`
5. Recursively check nested blocks (if/else, switch, loops, etc.)

## Configuration

Currently, the linter has no configuration options. It enforces the rule strictly: every `WriteHeader()` call must be immediately followed by a `return` statement.

## Contributing

This linter is designed for a specific use case. If you have suggestions for improvements or find bugs, please open an issue.

## License

MIT License - see LICENSE file for details
