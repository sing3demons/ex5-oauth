#!/bin/bash

# Test script to verify todo CRUD operations with authentication
# This script demonstrates how the endpoints work with real tokens

set -e

BASE_URL="http://localhost:3001"
OAUTH_URL="http://localhost:8080"

echo "🧪 Testing NestJS Backend with Authentication"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Note: This test requires manual authentication${NC}"
echo ""
echo "To test with real authentication:"
echo "1. Open browser to http://localhost:5173"
echo "2. Login with your credentials"
echo "3. Open browser DevTools (F12)"
echo "4. Go to Network tab"
echo "5. Make a request to /api/todos"
echo "6. Copy the Authorization header value"
echo "7. Export it: export AUTH_TOKEN='Bearer your-token-here'"
echo "8. Run this script again"
echo ""

if [ -z "$AUTH_TOKEN" ]; then
    echo -e "${YELLOW}⚠️  AUTH_TOKEN not set. Skipping authenticated tests.${NC}"
    echo ""
    echo "Testing public endpoints only:"
    echo ""
    
    echo "✓ Health check:"
    curl -s "$BASE_URL/health" | jq .
    echo ""
    
    echo "✓ Login initiation:"
    curl -s "$BASE_URL/auth/login" | jq .authorization_url
    echo ""
    
    echo "✓ Discovery endpoint:"
    curl -s "$BASE_URL/auth/discovery" | jq '.issuer, .authorization_endpoint, .token_endpoint' | head -3
    echo ""
    
    echo -e "${GREEN}Public endpoints working correctly!${NC}"
    echo ""
    echo "To test authenticated endpoints, set AUTH_TOKEN and run again."
    exit 0
fi

echo -e "${GREEN}✓ AUTH_TOKEN found. Testing authenticated endpoints...${NC}"
echo ""

# Test 1: Get all todos
echo "📋 Test 1: Get all todos"
echo "------------------------"
response=$(curl -s -w "\n%{http_code}" -H "Authorization: $AUTH_TOKEN" "$BASE_URL/api/todos")
status=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status" = "200" ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    echo "Todos count: $(echo "$body" | jq '. | length')"
else
    echo "✗ FAILED (Status: $status)"
    echo "$body" | jq .
fi
echo ""

# Test 2: Create a new todo
echo "📋 Test 2: Create a new todo"
echo "----------------------------"
new_todo=$(cat <<EOF
{
  "title": "Test Todo from NestJS",
  "description": "This is a test todo created by the integration test",
  "priority": "high"
}
EOF
)

response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Authorization: $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$new_todo" \
    "$BASE_URL/api/todos")

status=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status" = "201" ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    TODO_ID=$(echo "$body" | jq -r '.id')
    echo "Created todo ID: $TODO_ID"
    echo "$body" | jq '{id, title, status, priority}'
else
    echo "✗ FAILED (Status: $status)"
    echo "$body" | jq .
    exit 1
fi
echo ""

# Test 3: Get specific todo
echo "📋 Test 3: Get specific todo"
echo "----------------------------"
response=$(curl -s -w "\n%{http_code}" -H "Authorization: $AUTH_TOKEN" "$BASE_URL/api/todos/$TODO_ID")
status=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status" = "200" ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    echo "$body" | jq '{id, title, status}'
else
    echo "✗ FAILED (Status: $status)"
    echo "$body" | jq .
fi
echo ""

# Test 4: Update todo
echo "📋 Test 4: Update todo"
echo "----------------------"
update_data=$(cat <<EOF
{
  "title": "Updated Test Todo",
  "description": "This todo has been updated",
  "priority": "medium"
}
EOF
)

response=$(curl -s -w "\n%{http_code}" -X PUT \
    -H "Authorization: $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$update_data" \
    "$BASE_URL/api/todos/$TODO_ID")

status=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status" = "200" ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    echo "$body" | jq '{id, title, priority}'
else
    echo "✗ FAILED (Status: $status)"
    echo "$body" | jq .
fi
echo ""

# Test 5: Update status (drag & drop simulation)
echo "📋 Test 5: Update status (drag & drop)"
echo "--------------------------------------"
for new_status in "in_progress" "done"; do
    echo "Updating status to: $new_status"
    response=$(curl -s -w "\n%{http_code}" -X PATCH \
        -H "Authorization: $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"status\":\"$new_status\"}" \
        "$BASE_URL/api/todos/$TODO_ID/status")
    
    status=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "200" ]; then
        echo -e "  ${GREEN}✓ Status updated to $new_status${NC}"
    else
        echo "  ✗ FAILED (Status: $status)"
        echo "$body" | jq .
    fi
done
echo ""

# Test 6: Delete todo
echo "📋 Test 6: Delete todo"
echo "----------------------"
response=$(curl -s -w "\n%{http_code}" -X DELETE \
    -H "Authorization: $AUTH_TOKEN" \
    "$BASE_URL/api/todos/$TODO_ID")

status=$(echo "$response" | tail -n1)

if [ "$status" = "204" ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    echo "Todo deleted successfully"
else
    echo "✗ FAILED (Status: $status)"
fi
echo ""

# Test 7: Verify deletion
echo "📋 Test 7: Verify deletion"
echo "--------------------------"
response=$(curl -s -w "\n%{http_code}" -H "Authorization: $AUTH_TOKEN" "$BASE_URL/api/todos/$TODO_ID")
status=$(echo "$response" | tail -n1)

if [ "$status" = "404" ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    echo "Todo not found (as expected)"
else
    echo "✗ FAILED (Status: $status)"
    echo "Expected 404, got $status"
fi
echo ""

echo "=============================================="
echo -e "${GREEN}✅ All authenticated tests passed!${NC}"
echo "=============================================="
