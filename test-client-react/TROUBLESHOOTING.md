# Troubleshooting Guide

## ❌ "Failed to fetch" Error

### Problem
Browser shows "Failed to fetch" or CORS error in console.

### Solution
CORS middleware has been added to the OAuth server. Make sure you're running the latest version:

```bash
# Restart OAuth server
go run main.go
```

The server now includes CORS headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization, Accept`

### Verify CORS is Working

```bash
curl -X OPTIONS http://localhost:8080/oauth/token \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -v
```

Should see CORS headers in response.

## ❌ "No valid token available"

### Problem
App B or C shows "No valid token available. Please login first."

### Solution
1. Make sure you logged into App A first
2. Check browser console for errors
3. Check localStorage has tokens:
   ```javascript
   localStorage.getItem('sso_tokens')
   ```

## ❌ "Token exchange failed"

### Problem
Token exchange returns error.

### Causes & Solutions

### 1. Client credentials incorrect

Update `src/context/SSOContext.jsx` with correct client IDs and secrets from setup:

```javascript
const CLIENTS = {
  'app-a': {
    client_id: 'YOUR_ACTUAL_CLIENT_ID',
    client_secret: 'YOUR_ACTUAL_CLIENT_SECRET',
    // ...
  }
}
```

### 2. OAuth server not running

```bash
# Check if server is running
curl http://localhost:8080/health

# If not, start it
go run main.go
```

### 3. Token expired

Clear localStorage and login again:
```javascript
localStorage.clear()
```

## ❌ "Failed to get session ID"

### Problem
Login fails at authorization step.

### Solution

Check OAuth server logs for errors. Common issues:

1. **MongoDB not running**
   ```bash
   # Start MongoDB
   docker run -d -p 27017:27017 mongo:latest
   ```

2. **Database connection failed**
   Check `.env` file:
   ```
   MONGODB_URI=mongodb://localhost:27017
   DATABASE_NAME=oauth2_db
   ```

## ❌ "Invalid client credentials"

### Problem
Token exchange returns "invalid_client" error.

### Solution

Re-register clients:

```bash
cd test-client-react
./setup.sh
```

This will:
1. Register new clients
2. Update configuration automatically

## ❌ Network Errors

### Problem
All requests fail with network errors.

### Checklist

1. **OAuth server running?**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Correct port?**
   - OAuth server: `http://localhost:8080`
   - React app: `http://localhost:3000`

3. **Firewall blocking?**
   Check firewall settings

4. **Proxy issues?**
   Check `vite.config.js` proxy settings

## ❌ "User not found"

### Problem
Login fails with "Invalid email or password".

### Solution

Register user first:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

## 🔍 Debug Mode

### Enable Console Logging

Add to `src/context/SSOContext.jsx`:

```javascript
// At the top of each function
console.log('Login called with:', { appId, email })
console.log('Token exchange called for:', targetAppId)
console.log('Current tokens:', tokens)
```

### Check Network Tab

1. Open Browser DevTools (F12)
2. Go to Network tab
3. Try login
4. Check each request:
   - Status code
   - Request headers
   - Response body

### Common Status Codes

- `200 OK` - Success
- `400 Bad Request` - Invalid parameters
- `401 Unauthorized` - Invalid credentials
- `404 Not Found` - Endpoint not found
- `500 Internal Server Error` - Server error

## 🔧 Reset Everything

If nothing works, reset everything:

```bash
# 1. Stop OAuth server (Ctrl+C)

# 2. Stop React app (Ctrl+C)

# 3. Clear MongoDB
mongo
> use oauth2_db
> db.dropDatabase()
> exit

# 4. Clear browser data
# In browser: DevTools > Application > Clear storage

# 5. Restart OAuth server
go run main.go

# 6. Re-setup test client
cd test-client-react
./setup.sh
npm run dev
```

## 📞 Still Having Issues?

### Check These Files

1. **OAuth Server Logs**
   - Look for errors in terminal running `go run main.go`

2. **Browser Console**
   - F12 > Console tab
   - Look for red errors

3. **Network Tab**
   - F12 > Network tab
   - Check failed requests

### Common Patterns

#### Pattern 1: CORS Error
```
Access to fetch at 'http://localhost:8080/...' from origin 'http://localhost:3000' 
has been blocked by CORS policy
```
**Fix**: Restart OAuth server (CORS middleware added)

#### Pattern 2: 404 Not Found
```
POST http://localhost:8080/oauth/token 404 (Not Found)
```
**Fix**: Check OAuth server is running

#### Pattern 3: Invalid JSON
```
SyntaxError: Unexpected token < in JSON at position 0
```
**Fix**: Server returned HTML instead of JSON (probably an error page)

## ✅ Verification Checklist

Before testing, verify:

- [ ] OAuth server running on port 8080
- [ ] MongoDB running on port 27017
- [ ] React app running on port 3000
- [ ] User registered in OAuth server
- [ ] Clients registered (run setup.sh)
- [ ] Client credentials updated in SSOContext.jsx
- [ ] Browser console shows no CORS errors
- [ ] Can access http://localhost:8080/health

## 🎯 Quick Test

```bash
# Test OAuth server
curl http://localhost:8080/health
# Should return: OK

# Test CORS
curl -X OPTIONS http://localhost:8080/oauth/token \
  -H "Origin: http://localhost:3000" \
  -v
# Should see Access-Control-Allow-Origin header

# Test user registration
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'
# Should return success or "user exists"
```

If all tests pass, the system should work! 🎉
