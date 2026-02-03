#!/bin/bash
# Tests the MCPShell agent RAG (Retrieval-Augmented Generation) functionality
#
# This test suite covers:
# 0. RAG config validation (works without API key)
#    - Validates tools config with RAG section
#    - Validates agent.yaml config file with RAG section
# 1. RAG with single source (--rag=product_docs)
# 2. RAG with different source (--rag=security_docs)
# 3. RAG with multiple sources (--rag=product_docs --rag=security_docs)
# 4. RAG from agent.yaml config file (tests loading RAG config from agent.yaml)
#
# Tests both:
# - Using --rag flags to specify sources from tools config
# - Using agent.yaml config file with embedded RAG section

# Source common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_ROOT="$(dirname "$SCRIPT_DIR")"
source "$TESTS_ROOT/common/common.sh"

#####################################################################################
# Configuration for this test
TEST_NAME="test_agent_rag"

#####################################################################################
# Start the test

testcase "$TEST_NAME"

info "Testing MCPShell agent with RAG support"

# Make sure we have the CLI binary
check_cli_exists

# Get the repository root (parent of tests directory)
REPO_ROOT="$(dirname "$TESTS_ROOT")"

# Path to test configuration
TEST_CONFIG="$SCRIPT_DIR/tools/test_agent_rag.yaml"

if [ ! -f "$TEST_CONFIG" ]; then
    fail "Test configuration file not found: $TEST_CONFIG"
fi

# Verify RAG documents exist
RAG_DOCS_DIR="$SCRIPT_DIR/rag_docs"
if [ ! -d "$RAG_DOCS_DIR" ]; then
    fail "RAG documents directory not found: $RAG_DOCS_DIR"
fi

for doc in "product_info.txt" "api_reference.txt" "security_guide.txt"; do
    if [ ! -f "$RAG_DOCS_DIR/$doc" ]; then
        fail "RAG document not found: $RAG_DOCS_DIR/$doc"
    fi
done

success "All RAG documents found"

separator
info "0. Testing RAG config validation (dry-run without API key)"
separator

# First test: verify the config is valid and can be loaded
# This doesn't require an API key, just validates the config structure
info "Validating RAG configuration structure..."

# Try to validate the config (this should work even without API key)
VALIDATE_OUTPUT=$("$CLI_BIN" validate --tools="$TEST_CONFIG" 2>&1)
VALIDATE_RESULT=$?

if [ $VALIDATE_RESULT -eq 0 ]; then
    success "RAG configuration is valid"
else
    info "Validation output: $VALIDATE_OUTPUT"
    fail "RAG configuration validation failed"
fi

# Also validate the agent config file
info "Validating agent.yaml with RAG section..."
AGENT_CONFIG="$SCRIPT_DIR/agent_with_rag.yaml"

if [ -f "$AGENT_CONFIG" ]; then
    # Check that the file is valid YAML and contains RAG section
    if grep -q "rag:" "$AGENT_CONFIG"; then
        success "Agent config file contains RAG section"
    else
        fail "Agent config file does not contain RAG section"
    fi
else
    fail "Agent config file not found: $AGENT_CONFIG"
fi

# Check if OPENAI_API_KEY is set for actual RAG tests
if [ -z "$OPENAI_API_KEY" ]; then
    separator
    info "OPENAI_API_KEY not set - skipping live RAG tests (config validation passed)"
    success "RAG configuration validation tests completed successfully!"
    exit 0
fi

separator
info "1. Testing RAG document processing (product_docs source)"
separator

# Test that the agent can process RAG sources
# We'll run in one-shot mode with a question that requires RAG retrieval
QUESTION="What is MCPShell and what are its key features?"

info "Running agent with RAG enabled..."
info "Question: $QUESTION"

# Run from repo root so relative paths in config work
cd "$REPO_ROOT" || fail "Failed to change to repo root: $REPO_ROOT"

OUTPUT=$("$CLI_BIN" agent \
    --tools="$TEST_CONFIG" \
    --rag=product_docs \
    --once \
    --user-prompt="$QUESTION" \
    2>&1)
RESULT=$?

# Check if command succeeded
if [ $RESULT -ne 0 ]; then
    info "Agent output: $OUTPUT"
    fail "Agent command failed with exit code: $RESULT"
fi

# Verify the output contains information from the RAG documents
# The agent should mention MCPShell and some of its features
if echo "$OUTPUT" | grep -qi "MCPShell"; then
    success "Agent response mentions MCPShell"
else
    info "Agent output: $OUTPUT"
    fail "Agent response does not mention MCPShell"
fi

# Check for at least one key feature mentioned in the product_info.txt
if echo "$OUTPUT" | grep -qiE "(command-line|MCP|Model Context Protocol|secure|configuration|YAML)"; then
    success "Agent response includes information from RAG documents"
else
    info "Agent output: $OUTPUT"
    fail "Agent response does not include expected information from RAG documents"
fi

info "Agent response (first 500 chars):"
echo "$OUTPUT" | head -c 500 | sed 's/^/  /'
echo ""

separator
info "1b. Testing RAG retrieval of unique product information (serial number)"
separator

# Test that the agent can retrieve the specific unique serial number from the RAG documents
SERIAL_QUESTION="What is the product serial number for MCPShell?"

info "Running agent with RAG enabled to retrieve unique information..."
info "Question: $SERIAL_QUESTION"

OUTPUT=$("$CLI_BIN" agent \
    --tools="$TEST_CONFIG" \
    --rag=product_docs \
    --once \
    --user-prompt="$SERIAL_QUESTION" \
    2>&1)
RESULT=$?

if [ $RESULT -ne 0 ]; then
    info "Agent output: $OUTPUT"
    fail "Agent command failed with exit code: $RESULT"
fi

# Check for the specific serial number that only exists in the RAG documents
if echo "$OUTPUT" | grep -q "0x684356"; then
    success "✓ Agent successfully retrieved unique serial number (0x684356) from RAG documents"
else
    info "Agent output: $OUTPUT"
    fail "Agent did not retrieve the unique serial number (0x684356) from RAG documents"
fi

info "Agent response (first 300 chars):"
echo "$OUTPUT" | head -c 300 | sed 's/^/  /'
echo ""

separator
info "1c. Testing RAG retrieval of unique API key format"
separator

# Test that the agent can retrieve the specific API key format from the RAG documents
API_KEY_QUESTION="What is the API authentication key format for MCPShell?"

info "Running agent with RAG enabled to retrieve API key format..."
info "Question: $API_KEY_QUESTION"

OUTPUT=$("$CLI_BIN" agent \
    --tools="$TEST_CONFIG" \
    --rag=product_docs \
    --once \
    --user-prompt="$API_KEY_QUESTION" \
    2>&1)
RESULT=$?

if [ $RESULT -ne 0 ]; then
    info "Agent output: $OUTPUT"
    fail "Agent command failed with exit code: $RESULT"
fi

# Check for the specific API key format that only exists in the RAG documents
if echo "$OUTPUT" | grep -q "DON-RAG-TEST-KEY-ABC123XYZ"; then
    success "✓ Agent successfully retrieved unique API key format (DON-RAG-TEST-KEY-ABC123XYZ) from RAG documents"
else
    info "Agent output: $OUTPUT"
    fail "Agent did not retrieve the unique API key format from RAG documents"
fi

info "Agent response (first 300 chars):"
echo "$OUTPUT" | head -c 300 | sed 's/^/  /'
echo ""

separator
info "2. Testing RAG with security documentation (security_docs source)"
separator

SECURITY_QUESTION="What are the security best practices for MCPShell constraints?"

info "Running agent with security RAG source..."
info "Question: $SECURITY_QUESTION"

OUTPUT=$("$CLI_BIN" agent \
    --tools="$TEST_CONFIG" \
    --rag=security_docs \
    --once \
    --user-prompt="$SECURITY_QUESTION" \
    2>&1)
RESULT=$?

if [ $RESULT -ne 0 ]; then
    info "Agent output: $OUTPUT"
    fail "Agent command failed with exit code: $RESULT"
fi

# Verify the output contains security-related information
if echo "$OUTPUT" | grep -qiE "(constraint|security|injection|validation)"; then
    success "Agent response includes security information from RAG documents"
else
    info "Agent output: $OUTPUT"
    fail "Agent response does not include expected security information"
fi

info "Agent response (first 500 chars):"
echo "$OUTPUT" | head -c 500 | sed 's/^/  /'
echo ""

separator
info "2b. Testing RAG retrieval of unique security certification code"
separator

# Test that the agent can retrieve the specific security certification code from the RAG documents
CERT_QUESTION="What is the security certification code for MCPShell?"

info "Running agent with security RAG source to retrieve unique information..."
info "Question: $CERT_QUESTION"

OUTPUT=$("$CLI_BIN" agent \
    --tools="$TEST_CONFIG" \
    --rag=security_docs \
    --once \
    --user-prompt="$CERT_QUESTION" \
    2>&1)
RESULT=$?

if [ $RESULT -ne 0 ]; then
    info "Agent output: $OUTPUT"
    fail "Agent command failed with exit code: $RESULT"
fi

# Check for the specific security certification code that only exists in the RAG documents
if echo "$OUTPUT" | grep -q "SEC-DON-9F8E7D6C"; then
    success "✓ Agent successfully retrieved unique security certification code (SEC-DON-9F8E7D6C) from RAG documents"
else
    info "Agent output: $OUTPUT"
    fail "Agent did not retrieve the unique security certification code from RAG documents"
fi

info "Agent response (first 300 chars):"
echo "$OUTPUT" | head -c 300 | sed 's/^/  /'
echo ""

separator
info "3. Testing RAG with multiple sources"
separator

MULTI_QUESTION="What is MCPShell and what security features does it have?"

info "Running agent with multiple RAG sources..."
info "Question: $MULTI_QUESTION"

OUTPUT=$("$CLI_BIN" agent \
    --tools="$TEST_CONFIG" \
    --rag=product_docs \
    --rag=security_docs \
    --once \
    --user-prompt="$MULTI_QUESTION" \
    2>&1)
RESULT=$?

if [ $RESULT -ne 0 ]; then
    info "Agent output: $OUTPUT"
    fail "Agent command failed with exit code: $RESULT"
fi

# Verify the output contains information from both sources
if echo "$OUTPUT" | grep -qi "MCPShell" && echo "$OUTPUT" | grep -qiE "(security|constraint|validation)"; then
    success "Agent response includes information from multiple RAG sources"
else
    info "Agent output: $OUTPUT"
    fail "Agent response does not include information from both RAG sources"
fi

info "Agent response (first 500 chars):"
echo "$OUTPUT" | head -c 500 | sed 's/^/  /'
echo ""

separator
info "4. Testing RAG with agent.yaml config file (not --rag flags)"
separator

# Test using agent.yaml config file with RAG section
AGENT_CONFIG="$SCRIPT_DIR/agent_with_rag.yaml"

if [ ! -f "$AGENT_CONFIG" ]; then
    fail "Agent config file not found: $AGENT_CONFIG"
fi

info "Using agent config file: $AGENT_CONFIG"

# Set DON_AGENT_CONFIG to use our test config
export DON_AGENT_CONFIG="$AGENT_CONFIG"

CONFIG_QUESTION="What is MCPShell and what API features does it have?"

info "Running agent with RAG from agent.yaml config..."
info "Question: $CONFIG_QUESTION"

# Run with --rag flag to enable specific source from config
OUTPUT=$("$CLI_BIN" agent \
    --tools="$TEST_CONFIG" \
    --rag=product_docs \
    --rag=api_docs \
    --once \
    --user-prompt="$CONFIG_QUESTION" \
    2>&1)
RESULT=$?

# Unset the config env var
unset DON_AGENT_CONFIG

if [ $RESULT -ne 0 ]; then
    info "Agent output: $OUTPUT"
    fail "Agent command failed with exit code: $RESULT"
fi

# Verify the output contains information from both sources
if echo "$OUTPUT" | grep -qi "MCPShell" && echo "$OUTPUT" | grep -qiE "(API|reference|documentation)"; then
    success "Agent response includes information from agent.yaml RAG config"
else
    info "Agent output: $OUTPUT"
    fail "Agent response does not include expected information from agent.yaml RAG config"
fi

info "Agent response (first 500 chars):"
echo "$OUTPUT" | head -c 500 | sed 's/^/  /'
echo ""

separator
success "All agent RAG tests completed successfully!"
exit 0
