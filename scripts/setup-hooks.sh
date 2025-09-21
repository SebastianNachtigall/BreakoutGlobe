#!/bin/bash

# Setup script for Git hooks

echo "Setting up Git hooks..."

# Make sure hooks directory exists
mkdir -p .githooks

# Make hooks executable
chmod +x .githooks/*

# Configure Git to use our hooks directory
git config core.hooksPath .githooks

echo "✅ Git hooks configured successfully!"
echo ""
echo "The following hooks are now active:"
echo "  - pre-commit: Checks for escape character issues and runs tests"
echo ""
echo "To bypass hooks temporarily, use: git commit --no-verify"