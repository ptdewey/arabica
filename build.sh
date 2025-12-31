#!/usr/bin/env bash
set -e

echo "🔧 Building Arabica..."

# Generate templ files
echo "📝 Generating templates..."
templ generate

# Build CSS
echo "🎨 Building CSS..."
tailwindcss -i web/static/css/style.css -o web/static/css/output.css --minify

# Build Go binary
echo "🚀 Building Go application..."
mkdir -p bin
go build -o bin/arabica cmd/server/main.go

echo "✅ Build complete!"
echo ""
echo "Run './bin/arabica' to start the server"
echo "Or run 'make dev' for hot reload development mode"
