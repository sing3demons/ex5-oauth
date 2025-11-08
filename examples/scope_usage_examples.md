# OAuth Scope Usage Examples

## Why Scopes Matter

### Example 1: Photo Sharing App (Minimal Access)
```bash
# App เพียงแค่ต้องการ login
curl "http://localhost:8080/oauth/authorize?\
response_type=code&\
client_id=photo_app&\
scope=openid&\
redirect_uri=http://localhost:3000/callback"

# UserInfo Response
{
  "sub": "user123"
}
# ✅ App ได้แค่ user ID
# ❌ App ไม่ได้ email, name, phone
```

### Example 2: Email Newsletter App (Email Access)
```bash
# App ต้องการส่ง newsletter
curl "http://localhost:8080/oauth/authorize?\
response_type=code&\
client_id=newsletter_app&\
scope=openid%20email&\
redirect_uri=http://localhost:3000/callback"

# UserInfo Response
{
  "sub": "user123",
  "email": "user@example.com"
}
# ✅ App ได้ email เพื่อส่ง newsletter
# ❌ App ไม่ได้ name, phone
```

### Example 3: Profile Management App (Full Access)
```bash
# App ต้องการจัดการโปรไฟล์
curl "http://localhost:8080/oauth/authorize?\
response_type=code&\
client_id=profile_app&\
scope=openid%20profile%20email&\
redirect_uri=http://localhost:3000/callback"

# UserInfo Response
{
  "sub": "user123",
  "name": "John Doe",
  "email": "john@example.com"
}
# ✅ App ได้ทุกอย่าง
```

## Real-World Security Scenarios

### Scenario A: Token Leaked
```bash
# Token ที่มี scope จำกัด
Token: eyJ... (scope: openid email)

# Attacker ได้ token ไป
# ✅ สามารถ: ดู email
# ❌ ไม่สามารถ: แก้ไขโปรไฟล์, ลบบัญชี, เข้าถึงข้อมูลอื่น

# ความเสียหายจำกัด!
```

### Scenario B: Malicious App
```bash
# App ขอ scope มากเกินไป
scope=openid profile email phone address contacts calendar

# User เห็นแล้วสงสัย: "ทำไม calculator app ต้องการ contacts?"
# ❌ User ปฏิเสธ

# App ที่ดีขอเฉพาะที่จำเป็น
scope=openid

# ✅ User ไว้วางใจมากขึ้น
```

## API Protection Examples

### Protected Endpoint: Email API
```go
func SendEmailAPI(w http.ResponseWriter, r *http.Request) {
    token := extractToken(r)
    claims, _ := utils.ValidateToken(token, publicKey)
    
    // ต้องมี email scope
    if !utils.ScopeIncludesEmail(claims.Scope) {
        respondError(w, http.StatusForbidden, "insufficient_scope", 
            "This API requires 'email' scope")
        return
    }
    
    // ✅ มีสิทธิ์ - ดำเนินการต่อ
    sendEmail(claims.UserID, r.Body)
}
```

### Protected Endpoint: Profile Update API
```go
func UpdateProfileAPI(w http.ResponseWriter, r *http.Request) {
    token := extractToken(r)
    claims, _ := utils.ValidateToken(token, publicKey)
    
    // ต้องมี profile scope
    if !utils.ScopeIncludesProfile(claims.Scope) {
        respondError(w, http.StatusForbidden, "insufficient_scope", 
            "This API requires 'profile' scope")
        return
    }
    
    // ✅ มีสิทธิ์ - อัพเดทโปรไฟล์
    updateProfile(claims.UserID, r.Body)
}
```

## Scope Downgrade Example

### Use Case: Temporary Limited Access
```bash
# 1. ได้ token ที่มี scope เต็ม
POST /oauth/token
scope=openid profile email phone

# Response
{
  "access_token": "...",
  "scope": "openid profile email phone"
}

# 2. ต้องการ token ที่มี scope จำกัดกว่า (เช่น ส่งให้ third-party)
POST /token/exchange
subject_token=<full_scope_token>
scope=openid email
is_encrypted_jwe=true

# Response
{
  "access_token": "...",  # JWE token with limited scope
  "scope": "openid email"
}

# ✅ Third-party ได้แค่ email
# ❌ Third-party ไม่ได้ profile, phone
```

## Compliance Examples

### GDPR Compliance
```bash
# App ต้องบอก user ว่าเก็บข้อมูลอะไร

# Privacy Policy:
"เราเก็บข้อมูล:
- อีเมล (email scope) - เพื่อส่ง notification
- ชื่อ (profile scope) - เพื่อแสดงในแอป
เราไม่เก็บ: เบอร์โทร, ที่อยู่"

# OAuth Request ต้องตรงกับ Privacy Policy
scope=openid profile email  ✅
scope=openid profile email phone  ❌ (ไม่ได้บอกใน policy)
```

### Audit Log Example
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "event": "api_access",
  "user_id": "user123",
  "client_id": "app456",
  "endpoint": "/api/userinfo",
  "scope": "openid email",
  "ip": "192.168.1.1",
  "result": "success"
}
```

## Testing Scope Enforcement

### Test 1: Access Without Required Scope
```bash
# Get token with only openid scope
TOKEN=$(curl -s -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code&code=$CODE&..." \
  | jq -r .access_token)

# Try to access email API
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/email

# Expected Response:
{
  "error": "insufficient_scope",
  "error_description": "This API requires 'email' scope"
}
```

### Test 2: UserInfo Filtering
```bash
# Token with openid only
curl -H "Authorization: Bearer $TOKEN_OPENID" \
  http://localhost:8080/oauth/userinfo

# Response: {"sub":"user123"}

# Token with openid + email
curl -H "Authorization: Bearer $TOKEN_EMAIL" \
  http://localhost:8080/oauth/userinfo

# Response: {"sub":"user123","email":"user@example.com"}

# Token with openid + profile + email
curl -H "Authorization: Bearer $TOKEN_FULL" \
  http://localhost:8080/oauth/userinfo

# Response: {"sub":"user123","name":"John","email":"user@example.com"}
```

## Best Practices

### ✅ DO:
1. **Request minimal scopes**
   ```bash
   # Good
   scope=openid email
   ```

2. **Validate scope in every API**
   ```go
   if !utils.HasScope(claims.Scope, "email") {
       return error
   }
   ```

3. **Document required scopes**
   ```
   GET /api/email
   Required Scope: email
   ```

4. **Use scope downgrade for third-party**
   ```bash
   # Give limited token to third-party
   scope=openid  # minimal
   ```

### ❌ DON'T:
1. **Request all scopes**
   ```bash
   # Bad
   scope=openid profile email phone address contacts
   ```

2. **Ignore scope in API**
   ```go
   // Bad - no scope check
   func GetEmail(w http.ResponseWriter, r *http.Request) {
       return user.Email  // ❌ ไม่เช็ค scope
   }
   ```

3. **Use same token everywhere**
   ```bash
   # Bad - ใช้ token เดียวกันทุกที่
   # ถ้าหายเสียหายทั้งหมด
   ```

## Summary

Scope ช่วย:
1. 🔒 **Security** - จำกัดความเสียหายถ้า token หาย
2. 🔐 **Privacy** - User รู้ว่า app เข้าถึงอะไร
3. ⚖️ **Compliance** - ตาม GDPR, Privacy Laws
4. 🎯 **Least Privilege** - ให้สิทธิ์น้อยที่สุดที่จำเป็น
5. 📊 **Audit** - ตรวจสอบการใช้งานได้
6. 🛡️ **API Protection** - ป้องกัน unauthorized access

**หลักการ: "Ask for what you need, not what you want"**
