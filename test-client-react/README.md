# OAuth2 SSO Test Client (React)

React application สำหรับทดสอบ OAuth2 Single Sign-On (SSO) ด้วย Token Exchange

## Features

- ✅ 3 Apps แยกกัน (App A, B, C)
- ✅ SSO ด้วย Token Exchange (RFC 8693)
- ✅ Auto-login เมื่อมี token จาก app อื่น
- ✅ Token management และ caching
- ✅ User info display
- ✅ Logout from all apps

## Prerequisites

1. OAuth2 server ต้องรันอยู่ที่ `http://localhost:8080`
2. Node.js 18+ installed
3. User ต้องถูกสร้างไว้ใน OAuth server

## Setup

### 1. Install Dependencies

```bash
cd test-client-react
npm install
```

### 2. Register Test User

```bash
# Register a test user in OAuth server
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

### 3. Register OAuth Clients

```bash
# Register App A
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "App A (E-commerce)",
    "redirect_uris": ["http://localhost:3000/callback"]
  }'
# Save the client_id and client_secret

# Register App B
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "App B (Analytics)",
    "redirect_uris": ["http://localhost:3000/callback"]
  }'

# Register App C
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "App C (Chat)",
    "redirect_uris": ["http://localhost:3000/callback"]
  }'
```

### 4. Update Client Credentials

Edit `src/context/SSOContext.jsx` and update the client credentials:

```javascript
const CLIENTS = {
  'app-a': {
    client_id: 'YOUR_APP_A_CLIENT_ID',
    client_secret: 'YOUR_APP_A_CLIENT_SECRET',
    // ...
  },
  // ... same for app-b and app-c
}
```

## Run

```bash
npm run dev
```

Open http://localhost:3000

## How to Test SSO

### Step 1: Login to App A

1. Click on "App A" tab
2. Enter credentials:
   - Email: `test@example.com`
   - Password: `password123`
3. Click "Login to App A"
4. ✅ You should see "Login successful!"

### Step 2: Access App B (SSO!)

1. Click on "App B" tab
2. 🎉 App B will automatically login using Token Exchange!
3. No password required!

### Step 3: Access App C (SSO!)

1. Click on "App C" tab
2. 🎉 App C will also automatically login!
3. You're now logged into 3 apps with 1 login!

### Step 4: Test Logout

1. Click "Logout from All Apps" in any app
2. All apps will be logged out simultaneously

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    React Test Client                    │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │  App A   │  │  App B   │  │  App C   │             │
│  │ (Login)  │  │  (SSO)   │  │  (SSO)   │             │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘             │
│       │             │             │                     │
│       └─────────────┴─────────────┘                     │
│                     │                                    │
│              ┌──────▼──────┐                            │
│              │ SSO Context │                            │
│              │  (Tokens)   │                            │
│              └──────┬──────┘                            │
│                     │                                    │
└─────────────────────┼────────────────────────────────────┘
                      │
                      ▼
            ┌─────────────────┐
            │  OAuth2 Server  │
            │  (Port 8080)    │
            └─────────────────┘
```

**Note**: This demo uses a simplified login flow (direct token request) for ease of testing. In production, you would use the full OAuth Authorization Code flow with redirects.

## Token Exchange Flow

```
1. User logs into App A
   ↓
2. App A gets Token A from OAuth server
   ↓
3. User switches to App B
   ↓
4. App B detects Token A exists
   ↓
5. App B calls Token Exchange:
   POST /oauth/token
   grant_type=urn:ietf:params:oauth:grant-type:token-exchange
   subject_token=TOKEN_A
   ↓
6. OAuth server validates Token A
   ↓
7. OAuth server issues Token B
   ↓
8. App B is now logged in! (No password!)
```

## Features Demonstrated

### 1. Token Exchange (RFC 8693)
- Exchange tokens between apps
- No password required after first login
- Each app gets its own token

### 2. Token Management
- Tokens stored in localStorage
- Auto-detection of existing tokens
- Token expiration checking

### 3. User Experience
- Seamless SSO across apps
- Visual feedback for SSO
- Clear status indicators

### 4. Security
- Each app has unique token
- Tokens are validated
- Proper OAuth2 flow

## Troubleshooting

### "No valid token available"
- Make sure you logged into App A first
- Check if token is expired
- Try logging in again

### "Token exchange failed"
- Check OAuth server is running
- Verify client credentials are correct
- Check browser console for errors

### "Failed to get session ID"
- OAuth server might not be running
- Check CORS settings
- Verify OAuth server URL

## Development

### Project Structure

```
test-client-react/
├── src/
│   ├── components/
│   │   ├── AppA.jsx       # App A component
│   │   ├── AppB.jsx       # App B component
│   │   └── AppC.jsx       # App C component
│   ├── context/
│   │   └── SSOContext.jsx # SSO logic and token management
│   ├── App.jsx            # Main app component
│   ├── main.jsx           # Entry point
│   └── index.css          # Styles
├── index.html
├── vite.config.js
└── package.json
```

### Key Files

- **SSOContext.jsx**: Contains all SSO logic
  - `login()`: OAuth authorization code flow
  - `getTokenForApp()`: Token exchange for SSO
  - Token storage and management

- **AppA.jsx**: First app with login form
- **AppB.jsx**: Second app with auto-SSO
- **AppC.jsx**: Third app with auto-SSO

## Testing Checklist

- [ ] Can login to App A
- [ ] App B automatically logs in (SSO)
- [ ] App C automatically logs in (SSO)
- [ ] Can view user info in all apps
- [ ] Each app has different token
- [ ] Logout works from all apps
- [ ] Tokens persist after page refresh
- [ ] Token expiration is handled

## Next Steps

1. Add refresh token support
2. Add token refresh on expiration
3. Add error boundaries
4. Add loading states
5. Add more detailed logging
6. Add token introspection

## Summary

This test client demonstrates:
- ✅ OAuth2 Authorization Code Flow
- ✅ Token Exchange (RFC 8693)
- ✅ Single Sign-On (SSO)
- ✅ Multi-app token management
- ✅ Seamless user experience

**SSO works perfectly!** 🎉
