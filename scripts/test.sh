#!/bin/bash

# Test script for webpage-analyzer
# This script runs all unit tests and generates coverage reports

set -e

echo "🧪 Running unit tests for webpage-analyzer..."
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root."
    exit 1
fi

# Create coverage directory
mkdir -p coverage

print_status "Running tests for analyzer package..."
go test -v -coverprofile=coverage/analyzer.out -covermode=atomic ./internal/analyzer/...
if [ $? -eq 0 ]; then
    print_success "Analyzer tests passed"
else
    print_error "Analyzer tests failed"
    exit 1
fi

print_status "Running tests for HTTP handlers..."
go test -v -coverprofile=coverage/http.out -covermode=atomic ./internal/http/...
if [ $? -eq 0 ]; then
    print_success "HTTP handler tests passed"
else
    print_error "HTTP handler tests failed"
    exit 1
fi

print_status "Running tests for main application..."
go test -v -coverprofile=coverage/main.out -covermode=atomic ./cmd/webpage-analyzer/...
if [ $? -eq 0 ]; then
    print_success "Main application tests passed"
else
    print_error "Main application tests failed"
    exit 1
fi

print_status "Running all tests together..."
go test -v -coverprofile=coverage/all.out -covermode=atomic ./...
if [ $? -eq 0 ]; then
    print_success "All tests passed"
else
    print_error "Some tests failed"
    exit 1
fi

# Generate coverage reports
print_status "Generating coverage reports..."

echo "📊 Coverage Summary:"
echo "===================="

echo "Analyzer Package Coverage:"
go tool cover -func=coverage/analyzer.out

echo ""
echo "HTTP Handlers Coverage:"
go tool cover -func=coverage/http.out

echo ""
echo "Main Application Coverage:"
go tool cover -func=coverage/main.out

echo ""
echo "Overall Coverage:"
go tool cover -func=coverage/all.out

# Generate HTML coverage reports
print_status "Generating HTML coverage reports..."
go tool cover -html=coverage/analyzer.out -o coverage/analyzer.html
go tool cover -html=coverage/http.out -o coverage/http.html
go tool cover -html=coverage/main.out -o coverage/main.html
go tool cover -html=coverage/all.out -o coverage/all.html

print_success "HTML coverage reports generated in coverage/ directory"

# Run go vet
print_status "Running go vet..."
if go vet ./...; then
    print_success "go vet passed"
else
    print_error "go vet found issues"
    exit 1
fi

# Run go fmt check
print_status "Checking code formatting..."
if [ "$(gofmt -s -l . | wc -l)" -eq 0 ]; then
    print_success "Code formatting is correct"
else
    print_warning "Code formatting issues found. Run 'go fmt ./...' to fix."
    gofmt -s -l .
fi

# Run race detector (if supported)
print_status "Running race detector..."
if go test -race ./internal/analyzer/... 2>/dev/null; then
    print_success "Race detector passed for analyzer"
else
    print_warning "Race detector not available or found issues"
fi

if go test -race ./internal/http/... 2>/dev/null; then
    print_success "Race detector passed for HTTP handlers"
else
    print_warning "Race detector not available or found issues"
fi

echo ""
echo "🎉 All tests completed successfully!"
echo "📁 Coverage reports available in coverage/ directory"
echo "🌐 Open coverage/all.html in your browser to view detailed coverage"

# Optional: Open coverage report in browser (macOS)
if command -v open >/dev/null 2>&1; then
    read -p "Open coverage report in browser? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open coverage/all.html
    fi
fi 