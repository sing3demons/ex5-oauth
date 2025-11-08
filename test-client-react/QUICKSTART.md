# Quick Start Guide

## 🚀 เริ่มใช้งานใน 3 นาที!

### 1. Start OAuth Server

```bash
# ใน terminal แรก
cd /path/to/oauth-server
go run main.go
```

### 2. Setup Test Client

```bash
# ใน terminal ที่สอง
cd test-client-react
./setup.sh
npm install
npm run dev
```

### 3. Test SSO!

1. เปิด http://localhost:3000
2. Login ที่ **App A**:
   - Email: `test@example.com`
   - Password: `password123`
3. คลิกไปที่ **App B** → จะ login อัตโนมัติ! 🎉
4. คลิกไปที่ **App C** → จะ login อัตโนมัติ! 🎉

## ✨ สิ่งที่จะเห็น

### App A (First Login)
```
📱 App A - E-commerce
[Login Form]
↓
✅ Login successful!
```

### App B (Auto SSO)
```
📊 App B - Analytics
🔄 SSO in Progress...
↓
🎉 Automatically logged in via Token Exchange!
```

### App C (Auto SSO)
```
💬 App C - Chat
🔄 SSO in Progress...
↓
🎉 Automatically logged in via Token Exchange!
```

## 🎯 What's Happening?

1. **App A**: Normal OAuth login
   - User enters password
   - Gets Token A

2. **App B**: Token Exchange (SSO!)
   - Detects Token A exists
   - Exchanges Token A → Token B
   - No password needed!

3. **App C**: Token Exchange (SSO!)
   - Detects Token A or B exists
   - Exchanges → Token C
   - No password needed!

## 🔍 Behind the Scenes

```javascript
// App B automatically does this:
POST /oauth/token
{
  grant_type: "urn:ietf:params:oauth:grant-type:token-exchange",
  subject_token: "TOKEN_FROM_APP_A",
  subject_token_type: "urn:ietf:params:oauth:token-type:access_token",
  client_id: "app-b-client-id",
  client_secret: "app-b-secret"
}

// Response: New token for App B!
{
  access_token: "NEW_TOKEN_FOR_APP_B",
  token_type: "Bearer",
  expires_in: 3600
}
```

## 🎊 Result

**1 Login = 3 Apps Logged In!**

- ✅ App A: Logged in with password
- ✅ App B: Logged in via SSO (no password!)
- ✅ App C: Logged in via SSO (no password!)

## 🛠️ Troubleshooting

### OAuth server not running?
```bash
cd /path/to/oauth-server
go run main.go
```

### Setup failed?
```bash
# Make sure OAuth server is running first
curl http://localhost:8080/.well-known/openid-configuration

# Then run setup again
./setup.sh
```

### Login failed?
```bash
# Register user manually
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

## 📱 Features to Try

1. **Login to App A** → See success message
2. **Switch to App B** → Auto-login (SSO!)
3. **Switch to App C** → Auto-login (SSO!)
4. **View tokens** → Each app has different token
5. **Logout** → All apps logout together
6. **Refresh page** → Tokens persist (localStorage)

## 🎓 Learning Points

1. **Token Exchange** = SSO without cookies
2. **Each app** gets its own token
3. **No password** needed after first login
4. **Works with** SPA, Mobile, APIs
5. **Standards-compliant** (RFC 8693)

## 🚀 Ready!

Your OAuth2 SSO system is working! 🎉

Try it now: http://localhost:3000
