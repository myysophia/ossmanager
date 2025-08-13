#!/bin/bash

# WebDAV Proxy API Test Script
# This script demonstrates how to use the WebDAV proxy RESTful endpoints

# Configuration
BASE_URL="http://localhost:8080"
API_BASE="${BASE_URL}/api/v1/webdav/objects"
BUCKET="test-bucket"
JWT_TOKEN=""  # Set your JWT token here
WEBDAV_TOKEN=""  # Or set your WebDAV token here

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Function to make authenticated request
make_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local content_type="$4"
    
    local auth_header=""
    if [[ -n "$JWT_TOKEN" ]]; then
        auth_header="Authorization: Bearer $JWT_TOKEN"
    elif [[ -n "$WEBDAV_TOKEN" ]]; then
        auth_header="X-WebDAV-Token: $WEBDAV_TOKEN"
    else
        print_error "No authentication token provided"
        exit 1
    fi
    
    local cmd="curl -s -w '\n%{http_code}' -X $method"
    cmd+=" -H '$auth_header'"
    
    if [[ -n "$content_type" ]]; then
        cmd+=" -H 'Content-Type: $content_type'"
    fi
    
    if [[ -n "$data" ]]; then
        if [[ "$content_type" == *"json"* ]]; then
            cmd+=" -d '$data'"
        else
            cmd+=" --data-binary '$data'"
        fi
    fi
    
    cmd+=" '$url'"
    
    eval $cmd
}

# Check prerequisites
if [[ -z "$JWT_TOKEN" && -z "$WEBDAV_TOKEN" ]]; then
    print_error "Please set either JWT_TOKEN or WEBDAV_TOKEN in the script"
    echo "Example:"
    echo "JWT_TOKEN='your-jwt-token-here'"
    echo "WEBDAV_TOKEN='your-webdav-token-here'"
    exit 1
fi

print_header "WebDAV Proxy API Test"
echo "Testing WebDAV proxy endpoints at: $API_BASE"
echo "Bucket: $BUCKET"
echo ""

# Test 1: List root directory
print_header "Test 1: List Root Directory"
print_info "GET $API_BASE/$BUCKET"

response=$(make_request "GET" "$API_BASE/$BUCKET" "" "")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [[ "$http_code" == "200" ]]; then
    print_success "Directory listing successful"
    echo "$body" | jq -r '.data.items[] | "  - \(.name) (\(if .isDir then "directory" else "file, \(.size) bytes" end))"' 2>/dev/null || echo "$body"
else
    print_error "Directory listing failed (HTTP $http_code)"
    echo "$body"
fi
echo ""

# Test 2: Create a directory
print_header "Test 2: Create Directory"
print_info "POST $API_BASE/$BUCKET/mkdir"

mkdir_data='{"path":"/test-folder"}'
response=$(make_request "POST" "$API_BASE/$BUCKET/mkdir" "$mkdir_data" "application/json")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [[ "$http_code" == "200" ]]; then
    print_success "Directory created successfully"
    echo "$body" | jq -r '.data.message // .message' 2>/dev/null || echo "$body"
else
    print_error "Directory creation failed (HTTP $http_code)"
    echo "$body"
fi
echo ""

# Test 3: Upload a file
print_header "Test 3: Upload File"
print_info "POST $API_BASE/$BUCKET/file?path=/test-folder/test.txt"

# Create test content
test_content="Hello, WebDAV Proxy API!\nThis is a test file.\nTimestamp: $(date)"

response=$(make_request "POST" "$API_BASE/$BUCKET/file?path=/test-folder/test.txt" "$test_content" "text/plain")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [[ "$http_code" == "200" ]]; then
    print_success "File uploaded successfully"
    echo "$body" | jq -r '.data.message // .message' 2>/dev/null || echo "$body"
else
    print_error "File upload failed (HTTP $http_code)"
    echo "$body"
fi
echo ""

# Test 4: List directory with new content
print_header "Test 4: List Updated Directory"
print_info "GET $API_BASE/$BUCKET?path=/test-folder"

response=$(make_request "GET" "$API_BASE/$BUCKET?path=/test-folder" "" "")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [[ "$http_code" == "200" ]]; then
    print_success "Directory listing successful"
    echo "$body" | jq -r '.data.items[] | "  - \(.name) (\(if .isDir then "directory" else "file, \(.size) bytes" end))"' 2>/dev/null || echo "$body"
else
    print_error "Directory listing failed (HTTP $http_code)"
    echo "$body"
fi
echo ""

# Test 5: Rename file
print_header "Test 5: Rename File"
print_info "PATCH $API_BASE/$BUCKET/rename"

rename_data='{"oldPath":"/test-folder/test.txt","newPath":"/test-folder/renamed-test.txt"}'
response=$(make_request "PATCH" "$API_BASE/$BUCKET/rename" "$rename_data" "application/json")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [[ "$http_code" == "200" ]]; then
    print_success "File renamed successfully"
    echo "$body" | jq -r '.data.message // .message' 2>/dev/null || echo "$body"
else
    print_error "File rename failed (HTTP $http_code)"
    echo "$body"
fi
echo ""

# Test 6: Delete file
print_header "Test 6: Delete File"
print_info "DELETE $API_BASE/$BUCKET?path=/test-folder/renamed-test.txt"

response=$(make_request "DELETE" "$API_BASE/$BUCKET?path=/test-folder/renamed-test.txt" "" "")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [[ "$http_code" == "200" ]]; then
    print_success "File deleted successfully"
    echo "$body" | jq -r '.data.message // .message' 2>/dev/null || echo "$body"
else
    print_error "File deletion failed (HTTP $http_code)"
    echo "$body"
fi
echo ""

# Test 7: Delete directory
print_header "Test 7: Delete Directory"
print_info "DELETE $API_BASE/$BUCKET?path=/test-folder"

response=$(make_request "DELETE" "$API_BASE/$BUCKET?path=/test-folder" "" "")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [[ "$http_code" == "200" ]]; then
    print_success "Directory deleted successfully"
    echo "$body" | jq -r '.data.message // .message' 2>/dev/null || echo "$body"
else
    print_error "Directory deletion failed (HTTP $http_code)"
    echo "$body"
fi
echo ""

print_header "Test Complete"
print_info "All WebDAV proxy API tests completed!"
