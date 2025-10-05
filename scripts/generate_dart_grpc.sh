#!/bin/bash

# Script to generate Dart gRPC code from proto files
# This script should be run from the project root directory
#
# Requirements:
#   - protoc (Protocol Buffers compiler)
#   - flutter (Flutter SDK)
#   - protoc_plugin 21.1.2 (compatible with protobuf 4.x)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROTO_DIR="$PROJECT_ROOT/api/yao/oracle/v1"
DART_OUT_DIR="$PROJECT_ROOT/frontend/dashboard/lib/generated"
DASHBOARD_DIR="$PROJECT_ROOT/frontend/dashboard"

# Required protoc_plugin version (compatible with protobuf ^4.2.0)
REQUIRED_PROTOC_PLUGIN_VERSION="21.1.2"

echo "🔧 Generating Dart gRPC code..."
echo "Proto directory: $PROTO_DIR"
echo "Output directory: $DART_OUT_DIR"
echo ""

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "❌ Error: protoc is not installed"
    echo "Install it with:"
    echo "  macOS:   brew install protobuf"
    echo "  Linux:   apt-get install protobuf-compiler"
    echo "  Windows: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Check if flutter is installed
if ! command -v flutter &> /dev/null; then
    echo "❌ Error: flutter is not installed"
    echo "Install it from: https://flutter.dev/docs/get-started/install"
    exit 1
fi

# Ensure ~/.pub-cache/bin is in PATH
if [[ ":$PATH:" != *":$HOME/.pub-cache/bin:"* ]]; then
    echo "⚠️  Warning: ~/.pub-cache/bin is not in PATH"
    echo "Adding to PATH for this session..."
    export PATH="$PATH:$HOME/.pub-cache/bin"
fi

# Check if protoc-gen-dart is installed
if ! command -v protoc-gen-dart &> /dev/null; then
    echo "⚠️  protoc-gen-dart is not installed"
    echo "Installing protoc_plugin $REQUIRED_PROTOC_PLUGIN_VERSION..."
    flutter pub global activate protoc_plugin "$REQUIRED_PROTOC_PLUGIN_VERSION"
else
    # Check installed version
    INSTALLED_VERSION=$(flutter pub global list | grep protoc_plugin | awk '{print $2}' || echo "unknown")
    echo "📦 Current protoc_plugin version: $INSTALLED_VERSION"
    
    if [[ "$INSTALLED_VERSION" != "$REQUIRED_PROTOC_PLUGIN_VERSION" ]]; then
        echo "⚠️  Version mismatch detected!"
        echo "   Required: $REQUIRED_PROTOC_PLUGIN_VERSION"
        echo "   Installed: $INSTALLED_VERSION"
        echo "Installing correct version..."
        flutter pub global activate protoc_plugin "$REQUIRED_PROTOC_PLUGIN_VERSION"
    else
        echo "✅ Correct protoc_plugin version installed"
    fi
fi

# Create output directory if it doesn't exist
mkdir -p "$DART_OUT_DIR"

# Generate Dart code for dashboard.proto
echo ""
echo "📦 Generating Dart code for dashboard.proto..."
protoc \
    --dart_out=grpc:"$DART_OUT_DIR" \
    --proto_path="$PROJECT_ROOT/api" \
    yao/oracle/v1/dashboard.proto

echo "✅ Dart gRPC code generated successfully!"
echo "Output: $DART_OUT_DIR"

# Create .gitignore in generated directory
cat > "$DART_OUT_DIR/.gitignore" << EOF
# Generated gRPC code
*.pb.dart
*.pbgrpc.dart
*.pbenum.dart
*.pbjson.dart
EOF

# Verify generated files
echo ""
echo "📋 Generated files:"
ls -lh "$DART_OUT_DIR/yao/oracle/v1/"

# Run flutter pub get to update dependencies
echo ""
echo "📦 Running flutter pub get to update dependencies..."
cd "$DASHBOARD_DIR"
flutter pub get

echo ""
echo "✅ Done! Dart gRPC code generation complete."
echo ""
echo "Note: This script uses protoc_plugin $REQUIRED_PROTOC_PLUGIN_VERSION"
echo "      which is compatible with protobuf ^4.2.0 in pubspec.yaml"


