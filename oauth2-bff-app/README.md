# OAuth2 BFF Application (TypeScript)

Secure OAuth2 implementation using Backend-for-Frontend (BFF) pattern with maximum security.

## Architecture

```
React (TypeScript) → BFF Server (Node.js/TypeScript) → OAuth2 Server (Go)
```

## Security Features

✅ **HttpOnly Cookies** - Refresh tokens stored securely  
✅ **Confidential Client** - Client secret protected on server  
✅ **OIDC Compliant** - Full OpenID Connect implementation  
✅ **ID Token Validation** - Validates issuer, audience, nonce, expiry  
✅ **Token Rotation** - New refresh token on every refresh  
✅ **Memory-only Access Tokens** - No localStorage/sessionStorage  
✅ **CSRF Protection** - SameSite cookies + state parameter  
✅ **Nonce Protection** - Prevents replay attacks  
✅ **Auto Token Refresh** - Seamless user experience  

## Project Structure

```
oauth2-bff-app/
├── backend/              # BFF Server (TypeScript)
│   ├── src/
│   │   ├── server.ts
│   │   ├── routes/
│   │   ├── middleware/
│   │   └── types/
│   ├── package.json
│   └── tsconfig.json
│
└── frontend/            # React App (TypeScript)
    ├── src/
    │   ├── context/
    │   ├── hooks/
    │   ├── components/
    │   └── types/
    ├── package.json
    └── tsconfig.json
```

## Quick Start

### 1. Start OAuth2 Server (Go)
```bash
cd ../
go run main.go
```

### 2. Start BFF Server
```bash
cd backend
npm install
npm run dev
```

### 3. Start React Frontend
```bash
cd frontend
npm install
npm run dev
```

## Environment Variables

### Backend (.env)
```
PORT=3001
OAUTH2_SERVER_URL=http://localhost:8080
CLIENT_ID=qE5EjnNKrC6hRhYbC6q9VVND-rkN8Lah
CLIENT_SECRET=mfsw5Es8V0bSYrKYs3JCLlBYnIN322q2RlycNo3lLASnync03C2zYcDxXlLjwSXe
FRONTEND_URL=http://localhost:5173
SESSION_SECRET=your_secret_key
```

### Frontend (.env)
```
VITE_BFF_URL=http://localhost:3001
```

## How It Works

1. **Login Flow**:
   - User clicks login → BFF generates state for CSRF protection
   - Redirects to OAuth2 server
   - OAuth2 returns code → BFF exchanges for tokens (with client_secret)
   - BFF stores refresh token in HttpOnly cookie
   - Returns access token to frontend (memory only)

2. **Auto Refresh**:
   - Frontend detects token expiry
   - Calls BFF `/auth/refresh` endpoint
   - BFF uses HttpOnly cookie to refresh
   - Returns new access token

3. **Logout**:
   - Frontend calls BFF `/auth/logout`
   - BFF clears HttpOnly cookie
   - Frontend clears memory token

## Security Best Practices

- ✅ Refresh tokens never exposed to JavaScript
- ✅ Access tokens stored in memory only
- ✅ Client secret protected on server (never exposed to browser)
- ✅ HttpOnly + SameSite cookies prevent XSS/CSRF
- ✅ State parameter prevents CSRF attacks
- ✅ Token rotation on every refresh
- ✅ Short-lived access tokens (15 minutes)
- ✅ Long-lived refresh tokens (7 days)
