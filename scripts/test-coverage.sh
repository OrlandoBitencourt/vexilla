#!/bin/bash

echo "🧪 Running tests with coverage..."

go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

if [ $? -eq 0 ]; then
    echo "✅ All tests passed!"
    
    echo ""
    echo "📊 Coverage report:"
    go tool cover -func=coverage.out | tail -n 1
    
    echo ""
    echo "🌐 Generating HTML report..."
    go tool cover -html=coverage.out -o coverage.html
    
    echo "✅ Coverage report saved to coverage.html"
else
    echo "❌ Tests failed!"
    exit 1
fi
