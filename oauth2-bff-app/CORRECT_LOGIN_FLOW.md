# Correct OAuth2 Login Flow

## ⚠️ Common Mistake

**WRONG**: Accessing OAuth2 server directly
```
❌ http://localhost:8080/auth/login
```

**CORRECT**: Accessing Backend BFF
```
✅ http://localhost:4000/auth/login
```

## Why?

The OAuth2 flow in this application uses the **Backend-for-Frontend (BFF)** pattern:

```
Frontend → Backend BFF → OAuth2 Server
   ↓           ↓              ↓
  User    Handles Flow    Authenticates
```

### Flow Diagram

```
1. User clicks "Login" in Frontend
   ↓
2. Frontend calls: http://localhost:4000/auth/login
   ↓
3. Backend BFF generates PKCE parameters and returns authorization URL
   ↓
4. Frontend redirects user to OAuth2 Server
   ↓
5. User authenticates on OAuth2 Server
   ↓
6. OAuth2 Server redirects to: http://localhost:4000/auth/callback
   ↓
7. Backend BFF exchanges code for tokens
   ↓
8. Backend BFF redirects to Frontend with access_token
   ↓
9. Frontend stores token and user is logged in
```

## Correct Usage

### Option 1: Through Frontend (Recommended)

1. **Start Frontend**:
   ```bash
   cd oauth2-bff-app/frontend
   npm run dev
   ```

2. **Visit Frontend**:
   ```
   http://localhost:3000
   ```

3. **Click "Login with OAuth2"**:
   - Frontend will call backend
   - Backend will redirect to OAuth2 server
   - Complete authentication
   - You'll be redirected back to frontend

### Option 2: Direct API Call (For Testing)

1. **Call Backend BFF**:
   ```bash
   curl http://localhost:4000/auth/login
   ```

2. **Response**:
   ```json
   {
     "authorization_url": "http://localhost:8080/oauth/authorize?..."
   }
   ```

3. **Copy and Visit the authorization_url** in browser

4. **Complete OAuth2 flow**

## Error: "ไม่พบ redirect URI"

This error occurs when:

### Cause 1: Wrong URL

You accessed OAuth2 server directly instead of through Backend BFF:
```
❌ http://localhost:8080/auth/login
✅ http://localhost:4000/auth/login
```

### Cause 2: Client Not Registered

The OAuth2 client is not registered or has wrong redirect_uri.

**Solution**: Register the client

```bash
cd oauth2-bff-app
./register_todo_client.sh
```

**Expected redirect_uri**:
```
http://localhost:4000/auth/callback
```

### Cause 3: Wrong Redirect URI in Config

Check backend `.env`:

```bash
# oauth2-bff-app/backend/.env
OAUTH2_REDIRECT_URI=http://localhost:4000/auth/callback
```

## Port Reference

| Service | Port | URL |
|---------|------|-----|
| Frontend | 3000 | http://localhost:3000 |
| Backend BFF | 4000 | http://localhost:4000 |
| OAuth2 Server | 8080 | http://localhost:8080 |
| MongoDB | 27017 | mongodb://localhost:27017 |

## Correct Endpoints

### Frontend Endpoints
```
http://localhost:3000/          - Login page
http://localhost:3000/dashboard - Todo dashboard (protected)
http://localhost:3000/callback  - OAuth2 callback handler
```

### Backend BFF Endpoints
```
http://localhost:4000/auth/login     - Initiate OAuth2 flow
http://localhost:4000/auth/callback  - OAuth2 callback
http://localhost:4000/auth/refresh   - Refresh token
http://localhost:4000/auth/logout    - Logout
http://localhost:4000/api/todos      - Todos API
```

### OAuth2 Server Endpoints (Don't access directly!)
```
http://localhost:8080/oauth/authorize  - Authorization endpoint
http://localhost:8080/oauth/token      - Token endpoint
http://localhost:8080/oauth/userinfo   - User info endpoint
```

## Step-by-Step: Correct Login Flow

### 1. Start All Services

```bash
# Terminal 1: OAuth2 Server
./oauth2-server

# Terminal 2: MongoDB
mongod

# Terminal 3: Backend
cd oauth2-bff-app/backend
npm run dev

# Terminal 4: Frontend
cd oauth2-bff-app/frontend
npm run dev
```

### 2. Open Frontend

```
http://localhost:3000
```

### 3. Click "Login with OAuth2"

The frontend will:
1. Call `http://localhost:4000/auth/login`
2. Get authorization URL
3. Redirect you to OAuth2 server

### 4. Login on OAuth2 Server

Enter your credentials on the OAuth2 server page.

### 5. Automatic Redirect

After successful login:
1. OAuth2 server redirects to `http://localhost:4000/auth/callback`
2. Backend exchanges code for tokens
3. Backend redirects to `http://localhost:3000/callback` with access_token
4. Frontend stores token
5. You're redirected to dashboard

## Troubleshooting

### Issue: "ไม่พบ redirect URI"

**Check 1**: Are you using the correct URL?
```bash
# Wrong
curl http://localhost:8080/auth/login

# Correct
curl http://localhost:4000/auth/login
```

**Check 2**: Is the client registered?
```bash
cd oauth2-bff-app
./register_todo_client.sh
```

**Check 3**: Is the redirect_uri correct?
```bash
# Check backend .env
cat oauth2-bff-app/backend/.env | grep REDIRECT_URI

# Should be:
OAUTH2_REDIRECT_URI=http://localhost:4000/auth/callback
```

### Issue: "Connection Refused"

**Check**: Are all services running?
```bash
# Check OAuth2 server
curl http://localhost:8080/health

# Check Backend
curl http://localhost:4000/health

# Check Frontend
curl http://localhost:3000
```

### Issue: "CORS Error"

**Check**: Is CORS configured correctly?
```bash
# Backend .env should have:
FRONTEND_URL=http://localhost:3000
CORS_ORIGIN=http://localhost:3000
```

## Testing with curl

### 1. Get Authorization URL

```bash
curl http://localhost:4000/auth/login
```

**Response**:
```json
{
  "authorization_url": "http://localhost:8080/oauth/authorize?response_type=code&client_id=...&redirect_uri=http://localhost:4000/auth/callback&..."
}
```

### 2. Visit Authorization URL

Copy the `authorization_url` and paste it in your browser.

### 3. Complete Login

Login on the OAuth2 server page.

### 4. Get Tokens

After redirect, you'll have tokens in the URL or cookies.

## Summary

### ✅ DO:
- Access frontend at `http://localhost:3000`
- Use "Login with OAuth2" button
- Let the BFF handle OAuth2 flow
- Access backend at `http://localhost:4000/auth/login` for API testing

### ❌ DON'T:
- Access OAuth2 server directly at `http://localhost:8080/auth/login`
- Try to handle OAuth2 flow manually
- Skip the BFF layer

### Remember:
```
Frontend (3000) → Backend BFF (4000) → OAuth2 Server (8080)
```

The Backend BFF is the middleman that handles all OAuth2 complexity! 🎯
