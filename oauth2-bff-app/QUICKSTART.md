# Quick Start Guide

## Prerequisites

- Node.js 18+ installed
- OAuth2 Server running on `http://localhost:8080`
- Public client registered (from previous step)

## Setup

```bash
# Run setup script
./setup.sh

# Or manually:
cd backend && npm install && cp .env.example .env
cd ../frontend && npm install && cp .env.example .env
```

## Configuration

### Backend `.env`
```env
PORT=3001
OAUTH2_SERVER_URL=http://localhost:8080
CLIENT_ID=qE5EjnNKrC6hRhYbC6q9VVND-rkN8Lah
CLIENT_SECRET=mfsw5Es8V0bSYrKYs3JCLlBYnIN322q2RlycNo3lLASnync03C2zYcDxXlLjwSXe
FRONTEND_URL=http://localhost:5173
SESSION_SECRET=your-secret-key-here
NODE_ENV=development
```

### Frontend `.env`
```env
VITE_BFF_URL=http://localhost:3001
```

## Run

### Terminal 1: OAuth2 Server (Go)
```bash
cd ..
go run main.go
```

### Terminal 2: BFF Server
```bash
cd oauth2-bff-app/backend
npm run dev
```

### Terminal 3: React Frontend
```bash
cd oauth2-bff-app/frontend
npm run dev
```

## Test

1. Open http://localhost:5173
2. Click "Login with OAuth2"
3. Enter credentials on OAuth2 server
4. You'll be redirected back and logged in
5. **Try refreshing the page** - you stay logged in! ✨
6. **Try opening in another tab** - already logged in! ✨
7. **Logout in one tab** - all tabs logout! ✨

## How It Works

### Login Flow
```
1. User clicks "Login"
2. Frontend → BFF: GET /auth/login
3. BFF generates state for CSRF protection
4. BFF → Frontend: Returns authorization URL
5. Frontend redirects to OAuth2 server
6. User authenticates
7. OAuth2 → BFF: Callback with code
8. BFF exchanges code for tokens (with client_secret)
9. BFF stores refresh_token in HttpOnly cookie
10. BFF → Frontend: Redirects with access_token
11. Frontend stores access_token in memory
12. Frontend fetches user info
```

### Auto Refresh Flow
```
1. Frontend detects token will expire soon
2. Frontend → BFF: POST /auth/refresh (with HttpOnly cookie)
3. BFF uses refresh_token from cookie
4. BFF → OAuth2: Refresh token request
5. OAuth2 → BFF: New tokens
6. BFF updates HttpOnly cookie (token rotation)
7. BFF → Frontend: New access_token
8. Frontend updates memory token
9. Schedule next refresh
```

### Security Features

| Feature | Implementation | Benefit |
|---------|---------------|---------|
| **HttpOnly Cookies** | Refresh token stored in cookie | Cannot be accessed by JavaScript (XSS protection) |
| **Confidential Client** | Client secret on server only | Secret never exposed to browser |
| **State Parameter** | Random state for CSRF | Prevents CSRF attacks |
| **Memory-only Access Token** | Never stored in localStorage | Lost on page refresh, but auto-refreshed |
| **Token Rotation** | New refresh token on every refresh | Limits damage if token is compromised |
| **SameSite Cookies** | `SameSite=Lax` | Additional CSRF protection |
| **Auto Refresh** | Refresh before expiry | Seamless UX |
| **Multi-tab Sync** | localStorage events | Logout syncs across tabs |

## Troubleshooting

### "Failed to fetch"
- Check if BFF server is running on port 3001
- Check if OAuth2 server is running on port 8080

### "Invalid client"
- Make sure CLIENT_ID and CLIENT_SECRET in backend/.env match registered client
- Make sure client is registered as confidential client (with secret)

### "Refresh token expired"
- Refresh tokens expire after 7 days
- User needs to login again

### CORS errors
- Check FRONTEND_URL in backend/.env matches frontend URL
- Make sure withCredentials: true in axios calls

## Production Deployment

### Environment Variables

**Backend:**
```env
NODE_ENV=production
PORT=3001
OAUTH2_SERVER_URL=https://oauth.yourdomain.com
CLIENT_ID=your_production_client_id
CLIENT_SECRET=your_production_client_secret
FRONTEND_URL=https://app.yourdomain.com
SESSION_SECRET=use-a-strong-random-secret
```

**Frontend:**
```env
VITE_BFF_URL=https://bff.yourdomain.com
```

### Security Checklist

- [ ] Use HTTPS everywhere
- [ ] Set `secure: true` for cookies
- [ ] Use strong SESSION_SECRET and CLIENT_SECRET
- [ ] Store CLIENT_SECRET securely (environment variables, secrets manager)
- [ ] Enable rate limiting on BFF
- [ ] Use Redis for session storage (not in-memory)
- [ ] Set up monitoring and logging
- [ ] Implement additional CSRF tokens if needed
- [ ] Add request validation
- [ ] Set up proper CORS policies
- [ ] Use environment-specific configs
- [ ] Rotate CLIENT_SECRET periodically

## Architecture Benefits

✅ **Maximum Security** - Refresh tokens never exposed to JavaScript  
✅ **Great UX** - Auto-refresh, no repeated logins  
✅ **Scalable** - BFF can be scaled independently  
✅ **Flexible** - Easy to add more OAuth providers  
✅ **Maintainable** - Clear separation of concerns  
✅ **Mobile-ready** - Same BFF can serve mobile apps  

## Next Steps

- Add more OAuth providers (Google, GitHub, etc.)
- Implement token introspection
- Add user profile management
- Implement consent management
- Add session management UI
- Set up monitoring and analytics
