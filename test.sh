#!/bin/bash

# Integration test for Cloudless
# This script tests the basic functionality of Cloudless

set -e

echo "=== Cloudless Integration Test ==="
echo ""

# Build the binary
echo "1. Building cloudless..."
go build -o /tmp/cloudless ./cmd/cloudless
CLOUDLESS="/tmp/cloudless"

# Test 1: Check help command
echo "2. Testing help command..."
$CLOUDLESS --help > /dev/null
echo "   ✓ Help command works"

# Test 2: Check daemon status when not running
echo "3. Testing daemon status (should be not running)..."
OUTPUT=$($CLOUDLESS daemon status)
if [[ $OUTPUT == *"not running"* ]]; then
    echo "   ✓ Daemon status check works"
else
    echo "   ✗ Unexpected daemon status: $OUTPUT"
    exit 1
fi

# Test 3: List services when none exist
echo "4. Testing service list (should be empty)..."
OUTPUT=$($CLOUDLESS service list)
if [[ $OUTPUT == *"No services found"* ]]; then
    echo "   ✓ Service list works when empty"
else
    echo "   ✗ Unexpected service list output: $OUTPUT"
    exit 1
fi

# Test 4: Try to create service without daemon (should fail)
echo "5. Testing service creation without daemon (should fail)..."
if $CLOUDLESS service create postgres test-db 2>&1 | grep -q "daemon is not running"; then
    echo "   ✓ Correctly prevents service creation without daemon"
else
    echo "   ✗ Service creation should have failed without daemon"
    exit 1
fi

# Test 5: Verify Docker is available
echo "6. Verifying Docker is available..."
if docker ps > /dev/null 2>&1; then
    echo "   ✓ Docker is available"
else
    echo "   ⚠ Docker is not available - skipping container tests"
    echo ""
    echo "=== Integration Test Completed (Partial) ==="
    exit 0
fi

echo ""
echo "=== All Tests Passed ==="
echo ""
echo "Note: Full integration tests with daemon require running Docker and"
echo "cannot be fully automated in this environment. Manual testing required for:"
echo "  - Starting the daemon"
echo "  - Creating PostgreSQL services"
echo "  - Health checks and auto-restart"
echo "  - Service lifecycle management"
