#!/bin/bash

# Run all Don tests
# This script runs all test files in the tests directory

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Source common functions
source common/common.sh

info "Running Don tests..."

# Find and run all test scripts
TEST_SCRIPTS=$(find . -name "test_*.sh" -type f | sort)

if [ -z "$TEST_SCRIPTS" ]; then
    info "No test scripts found"
    exit 0
fi

PASSED=0
FAILED=0
SKIPPED=0

for script in $TEST_SCRIPTS; do
    info "Running $script..."
    
    if bash "$script"; then
        success "$script passed"
        ((PASSED++))
    else
        exit_code=$?
        if [ $exit_code -eq 77 ]; then
            info "$script skipped"
            ((SKIPPED++))
        else
            fail "$script failed with exit code $exit_code"
            ((FAILED++))
        fi
    fi
    
    echo ""
done

info "Test Summary:"
info "  Passed:  $PASSED"
info "  Failed:  $FAILED"
info "  Skipped: $SKIPPED"

if [ $FAILED -gt 0 ]; then
    fail "Some tests failed"
    exit 1
fi

success "All tests passed"
exit 0

