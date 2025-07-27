#!/bin/bash

set -e

echo "🔍 Running static code analysis..."

# Install golangci-lint if not present
if ! command -v golangci-lint &> /dev/null; then
    echo "📦 Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Run golangci-lint
echo "🔍 Running golangci-lint..."
golangci-lint run --timeout=5m --skip-dirs=api --skip-files=".*docs\\.go$"

# Run go vet
echo "🔍 Running go vet..."
go vet ./...

# Run go fmt check
echo "🔍 Checking code formatting..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    echo "❌ Code is not formatted properly. Run 'go fmt ./...' to fix."
    gofmt -s -l .
    exit 1
fi

# Run go mod tidy check
echo "🔍 Checking go.mod consistency..."
go mod tidy
if [ -n "$(git status --porcelain go.mod go.sum 2>/dev/null || echo '')" ]; then
    echo "❌ go.mod or go.sum has uncommitted changes. Run 'go mod tidy' and commit changes."
    exit 1
fi

# Run staticcheck
echo "🔍 Running staticcheck..."
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
else
    echo "📦 Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@v0.4.6
    staticcheck ./...
fi

echo "✅ All static analysis checks passed!" 