# ทดสอบ OAuth2 Flow ผ่าน Browser

## ขั้นตอนการทดสอบ

### 1. เริ่มเซิร์ฟเวอร์

```bash
go run main.go
```

### 2. ลงทะเบียน OAuth Client

```bash
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Application",
    "redirect_uris": ["http://localhost:3000/callback"]
  }'
```

บันทึก `client_id` และ `client_secret` ที่ได้

### 3. เริ่ม Authorization Flow

เปิด browser และไปที่:

```
http://localhost:8080/oauth/authorize?response_type=code&client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=random123
```

**สิ่งที่จะเกิดขึ้น:**
1. ระบบจะสร้าง session และ redirect ไปหน้า login
2. URL จะเป็น: `http://localhost:8080/auth/login?session_id=XXXXX`

### 4. Login หรือ Register

**กรณีที่ 1: ยังไม่มีบัญชี**
- คลิก "ลงทะเบียน" ที่หน้า login
- กรอกข้อมูล: ชื่อ, อีเมล, รหัสผ่าน
- กด "ลงทะเบียน"
- ระบบจะ:
  - สร้างบัญชีผู้ใช้
  - สร้าง authorization code (รูปแบบ: `{random}_{session_id}`)
  - Redirect กลับไปที่ `redirect_uri` พร้อม code

**กรณีที่ 2: มีบัญชีแล้ว**
- กรอกอีเมลและรหัสผ่าน
- กด "เข้าสู่ระบบ"
- ระบบจะ:
  - ตรวจสอบ credentials
  - สร้าง authorization code (รูปแบบ: `{random}_{session_id}`)
  - Redirect กลับไปที่ `redirect_uri` พร้อม code

### 5. Callback URL

หลังจาก login สำเร็จ browser จะ redirect ไปที่:

```
http://localhost:3000/callback?code=abc123_sessionid456&state=random123
```

**Authorization Code Format:**
- รูปแบบ: `{random_16_chars}_{session_id}`
- ตัวอย่าง: `a1b2c3d4e5f6g7h8_9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x`
- ประโยชน์: สามารถ trace กลับไปหา session ได้ง่าย สำหรับ logging และ debugging

### 6. แลก Code เป็น Token

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=YOUR_CODE&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&redirect_uri=http://localhost:3000/callback"
```

จะได้:
```json
{
  "access_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGc...",
  "id_token": "eyJhbGc...",
  "scope": "openid profile email"
}
```

## Flow Diagram

```
1. Client App
   ↓ (redirect to authorize)
2. OAuth Server: /oauth/authorize
   ↓ (create session, redirect to login)
3. Login Page: /auth/login?session_id=XXX
   ↓ (user enters credentials)
4. POST /auth/login
   ↓ (validate, create code with session_id)
5. Redirect to callback
   ↓ (code=random_sessionid)
6. Client App receives code
   ↓ (exchange code for tokens)
7. POST /oauth/token
   ↓ (validate code, return tokens)
8. Client App has tokens
```

## ข้อดีของการใช้ Session

1. **Stateful Authorization**: เก็บสถานะระหว่าง authorize request และ login
2. **Security**: ตรวจสอบว่า login มาจาก authorize request ที่ถูกต้อง
3. **Traceability**: Code มี session_id ทำให้ trace ได้ง่าย
4. **User Experience**: ไม่ต้องส่ง parameters ซ้ำๆ ระหว่าง login และ register

## การ Debug

ดู logs ใน MongoDB:

```javascript
// ดู sessions
db.sessions.find().pretty()

// ดู authorization codes
db.auth_codes.find().pretty()

// ดู users
db.users.find().pretty()
```

## หมายเหตุ

- Session จะหมดอายุใน 10 นาที
- Authorization code จะหมดอายุใน 10 นาที
- Code สามารถใช้ได้ครั้งเดียว (one-time use)
- หลังใช้ code แล้วจะถูกลบออกจาก database
