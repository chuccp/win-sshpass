#!/bin/bash

# Build all documentation sites

echo "Building Chinese (Simplified) documentation..."
mkdocs build -f mkdocs-zh.yml

echo "Building English documentation..."
mkdocs build -f mkdocs-en.yml

echo "Building Japanese documentation..."
mkdocs build -f mkdocs-ja.yml

echo "Building Chinese (Traditional) documentation..."
mkdocs build -f mkdocs-tw.yml

echo "Done!"
echo ""
echo "Chinese (Simplified): site-zh/"
echo "English:              site-en/"
echo "Japanese:             site-ja/"
echo "Chinese (Traditional): site-tw/"
echo ""
echo "To serve locally:"
echo "  mkdocs serve -f mkdocs-zh.yml  (Chinese Simplified)"
echo "  mkdocs serve -f mkdocs-en.yml  (English)"
echo "  mkdocs serve -f mkdocs-ja.yml  (Japanese)"
echo "  mkdocs serve -f mkdocs-tw.yml  (Chinese Traditional)"
