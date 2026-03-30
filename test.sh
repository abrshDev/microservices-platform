#!/bin/bash

# Configuration - Generating unique credentials using timestamp
TS=$(date +%s)
API_URL="http://localhost:8080/api/v1"
EMAIL="dev_$TS@abrham.et"
PASSWORD="Pass_$TS"
USERNAME="user_$TS"

echo "Starting Dynamic API Integration Tests (ID: $TS)"
echo "----------------------------------------------------------"

# 1. Register a fresh user
echo "1. Registering user ($EMAIL)..."
REGISTER_RES=$(curl -s -X POST "$API_URL/users" \
     -H "Content-Type: application/json" \
     -d "{\"username\": \"$USERNAME\", \"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")
echo "Response: $REGISTER_RES"
echo ""

# 2. Login & Capture Token
echo "2. Logging in..."
LOGIN_RES=$(curl -s -X POST "$API_URL/login" \
     -H "Content-Type: application/json" \
     -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

# Extract Token (Looking for the JWT string starting with eyJ)
TOKEN=$(echo $LOGIN_RES | grep -oP 'eyJ[a-zA-Z0-9\._\-]+')
# Extract User ID from the response body
USER_ID=$(echo $LOGIN_RES | grep -oP '(?<="user_id":")[^"]+')

if [ -z "$TOKEN" ]; then
    echo "Login Failed! No token received."
    echo "Full Response: $LOGIN_RES"
    exit 1
fi
echo "Login Successful! Token captured."
echo ""

# 3. Get User Profile (Passing the Bearer Token)
echo "3. Fetching user profile (Protected)..."
GET_RES=$(curl -s -X GET "$API_URL/users/$USER_ID" \
     -H "Authorization: Bearer $TOKEN")
echo "Response: $GET_RES"
echo ""

# 4. Delete User (Cleanup)
echo "4. Deleting user (Protected)..."
DELETE_RES=$(curl -s -X DELETE "$API_URL/users/$USER_ID" \
     -H "Authorization: Bearer $TOKEN")
echo "Response: $DELETE_RES"

echo "----------------------------------------------------------"
echo "Test Cycle for $EMAIL Completed!"