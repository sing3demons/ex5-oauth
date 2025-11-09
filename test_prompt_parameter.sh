#!/bin/bash

# Test script for OIDC prompt parameter support
# This script demonstrates the different prompt parameter behaviors

BASE_URL="http://localhost:8080"
CLIENT_ID="test-client"
REDIRECT_URI="http://localhost:3000/callback"
SCOPE="openid profile email"
STATE="test-state-123"

echo "=== OIDC Prompt Parameter Test Script ==="
echo ""
echo "This script demonstrates the different prompt parameter behaviors:"
echo "1. prompt=none - Fails if not authenticated or no consent"
echo "2. prompt=login - Forces re-authentication"
echo "3. prompt=consent - Forces consent screen"
echo "4. prompt=select_account - Forces account selection (placeholder)"
echo ""

# Test 1: prompt=none without authentication
echo "Test 1: prompt=none without SSO session"
echo "Expected: Redirect with error=login_required"
echo "URL: ${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=none"
echo ""

# Test 2: prompt=login
echo "Test 2: prompt=login (forces re-authentication)"
echo "Expected: Redirect to login page even if SSO session exists"
echo "URL: ${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=login"
echo ""

# Test 3: prompt=consent
echo "Test 3: prompt=consent (forces consent screen)"
echo "Expected: Show consent screen even if consent already granted"
echo "URL: ${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=consent"
echo ""

# Test 4: prompt=select_account
echo "Test 4: prompt=select_account (account selection)"
echo "Expected: Redirect to login page (placeholder behavior)"
echo "URL: ${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=select_account"
echo ""

echo "=== Test URLs Generated ==="
echo "Copy and paste these URLs in your browser to test:"
echo ""
echo "1. prompt=none:"
echo "${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=none"
echo ""
echo "2. prompt=login:"
echo "${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=login"
echo ""
echo "3. prompt=consent:"
echo "${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=consent"
echo ""
echo "4. prompt=select_account:"
echo "${BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=${SCOPE}&state=${STATE}&prompt=select_account"
