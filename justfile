set export

default:
    @just --list

# Build tool. Uses default if no tool is specified
build TOOL="":
    nix build{{ if TOOL == "" { "" } else { " .#" + TOOL } }}

# Run tool. Uses default if no tool is specified
run TOOL="":
    nix run{{ if TOOL == "" { "" } else { " .#" + TOOL } }}

# Run tests for specific tool
test TOOL:
    cd tools/{{TOOL}} && go test ./...

# Run tests for all tools
test-all:
    for dir in tools/*; do (cd "$dir" && go test ./...); done

# Run linter for specific tool
lint TOOL:
    cd tools/{{TOOL}} && golangci-lint run

# Run linter for all tools
lint-all:
    for dir in tools/*; do (cd "$dir" && golangci-lint run); done

# Clean build artifacts
clean:
    rm -rf result

# Update flake.lock
update:
    nix flake update

# Run pre-commit hooks
pre-commit:
    nix develop --command bash -c 'pre-commit run --all-files'

# Initialize a new tool
new-tool TOOL:
    mkdir -p tools/{{TOOL}}
    cd tools/{{TOOL}} && go mod init github.com/yourusername/ai-toolbox/tools/{{TOOL}}
    printf 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello, World from {{TOOL}}!") //nolint:forbidigo\n}\n' > tools/{{TOOL}}/main.go
    go work use ./tools/{{TOOL}}

