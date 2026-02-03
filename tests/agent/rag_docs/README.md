# RAG Test Documents

This directory contains static documents used for testing MCPShell's RAG (Retrieval-Augmented Generation) functionality.

## Documents

### product_info.txt
Contains general product information about MCPShell including:
- Product name, version, and category
- **Unique test value: Product Serial Number `0x684356`** (used to verify RAG retrieval)
- Description and key features
- Supported platforms
- License and contact information

### api_reference.txt
Contains API reference documentation including:
- **Unique test value: API Authentication Key Format `MCPSHELL-RAG-TEST-KEY-ABC123XYZ`** (used to verify RAG retrieval)
- Command-line interface commands and flags
- Configuration file format
- Constraint expression syntax
- Examples of common operations

### security_guide.txt
Contains security guidelines and best practices including:
- **Unique test value: Security Certification Code `SEC-MCPSHELL-9F8E7D6C`** (used to verify RAG retrieval)
- Security principles (defense in depth, least privilege, input validation)
- Constraint best practices for preventing common attacks
- Sandboxed runner configuration
- Environment variable handling
- Security checklist

## Usage in Tests

These documents are referenced in `tests/agent/tools/test_agent_rag.yaml` and used by `tests/agent/test_agent_rag.sh` to verify that:

1. RAG sources can be configured and loaded
2. Documents are processed correctly (downloaded/scanned)
3. The agent can retrieve relevant information from the documents
4. Multiple RAG sources can be used simultaneously
5. The agent provides accurate answers based on the retrieved content

## Test Scenarios

The test script (`test_agent_rag.sh`) runs the following scenarios:

1. **Product Documentation Query**: Tests retrieval from `product_docs` RAG source
   - Question: "What is MCPShell and what are its key features?"
   - Expected: Response includes information from product_info.txt and api_reference.txt

2. **Unique Serial Number Retrieval**: Tests RAG retrieval of specific unique information
   - Question: "What is the product serial number for MCPShell?"
   - Expected: Response includes the unique serial number `0x684356` from product_info.txt
   - **This verifies that RAG is actually retrieving information from the documents**

3. **Unique API Key Format Retrieval**: Tests RAG retrieval of specific unique information
   - Question: "What is the API authentication key format for MCPShell?"
   - Expected: Response includes the unique key format `MCPSHELL-RAG-TEST-KEY-ABC123XYZ` from api_reference.txt
   - **This verifies that RAG is actually retrieving information from the documents**

4. **Security Documentation Query**: Tests retrieval from `security_docs` RAG source
   - Question: "What are the security best practices for MCPShell constraints?"
   - Expected: Response includes information from security_guide.txt

5. **Unique Security Certification Code Retrieval**: Tests RAG retrieval of specific unique information
   - Question: "What is the security certification code for MCPShell?"
   - Expected: Response includes the unique code `SEC-MCPSHELL-9F8E7D6C` from security_guide.txt
   - **This verifies that RAG is actually retrieving information from the documents**

6. **Multi-Source Query**: Tests retrieval from multiple RAG sources
   - Question: "What is MCPShell and what security features does it have?"
   - Expected: Response includes information from both product_docs and security_docs

## Requirements

The RAG test requires:
- `OPENAI_API_KEY` environment variable set
- Internet connectivity (for OpenAI API calls)
- The test will be skipped if the API key is not available

## Adding New Test Documents

To add new test documents:

1. Create a new `.txt` file in this directory with relevant content
2. Update `tests/agent/tools/test_agent_rag.yaml` to include the new document in a RAG source
3. Update `tests/agent/test_agent_rag.sh` to verify the document content is retrieved correctly
4. Update this README to document the new document and test scenario
