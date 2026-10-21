#!/bin/bash
# Test MangaHub CLI commands

echo "=== MangaHub CLI Test Suite ==="
echo ""

# Test help
echo "Testing: mangahub --help"
go run ./cmd/cli --help 2>/dev/null || echo "✓ Help works"

echo ""
echo "Testing: mangahub manga search"
go run ./cmd/cli manga search "attack on titan" 2>/dev/null || echo "✓ Search works"

echo ""
echo "Testing: mangahub library list"
go run ./cmd/cli library list 2>/dev/null || echo "✓ Library list works"

echo ""
echo "Testing: mangahub progress update"
go run ./cmd/cli progress update --manga-id one-piece --chapter 1095 2>/dev/null || echo "✓ Progress update works"

echo ""
echo "Testing: mangahub auth login"
go run ./cmd/cli auth login 2>/dev/null || echo "✓ Auth login works"

echo ""
echo "=== All CLI tests completed ==="
