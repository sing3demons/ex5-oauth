# คู่มือการติดตั้งและทดสอบ

## ขั้นตอนการติดตั้ง

### 1. ติดตั้ง Dependencies

```bash
go mod download
```

### 2. เริ่ม MongoDB

ใช้ Docker:
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

หรือใช้ Makefile:
```bash
make docker-mongo
```

### 3. สร้างไฟล์ .env

```bash
cp .env.example .env
```

### 4. รันเซิร์ฟเวอร์

```bash
go run main.go
```

หรือ:
```bash
make run
```

เซิร์ฟเวอร์จะ:
- สร้าง RSA key pair อัตโนมัติ (2048-bit) ในครั้งแรก
- เก็บ keys ไว้ใน `keys/private.pem` และ `keys/public.pem`
- ใช้ keys เดิมในการรันครั้งต่อไป

## การทดสอบ API

### วิธีที่ 1: ใช้ Test Script

```bash
chmod +x test_api.sh
./test_api.sh
```

### วิธีที่ 2: ทดสอบด้วยตนเอง

#### 1. ลงทะเบียนผู้ใช้

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

#### 2. Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

จะได้ `access_token` ที่ sign ด้วย RS256

#### 3. ดู JWKS (Public Key)

```bash
curl http://localhost:8080/.well-known/jwks.json
```

#### 4. ดู OIDC Discovery

```bash
curl http://localhost:8080/.well-known/openid-configuration
```

## การตรวจสอบ JWT Token

คุณสามารถตรวจสอบ JWT token ที่ได้รับได้ที่ [jwt.io](https://jwt.io)

Token จะมี:
- Algorithm: RS256
- Header: `{"alg":"RS256","typ":"JWT"}`
- Payload: มี claims ต่างๆ เช่น sub, email, name, exp, iat

## โครงสร้าง Keys

```
keys/
├── private.pem  # RSA Private Key (2048-bit) - ใช้สำหรับ sign tokens
└── public.pem   # RSA Public Key - ใช้สำหรับ verify tokens
```

**สำคัญ:** 
- `private.pem` ต้องเก็บเป็นความลับ
- `public.pem` สามารถแชร์ได้ผ่าน JWKS endpoint
- ไม่ควร commit keys เข้า git (มี .gitignore แล้ว)

## Production Deployment

สำหรับ production:

1. สร้าง keys ล่วงหน้าด้วย:
```bash
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

2. เก็บ private key ใน secure storage (เช่น AWS Secrets Manager, HashiCorp Vault)

3. ตั้งค่า environment variables:
```bash
MONGODB_URI=mongodb://your-production-db:27017
DATABASE_NAME=oauth2_prod
SERVER_PORT=8080
ACCESS_TOKEN_EXPIRY=3600
REFRESH_TOKEN_EXPIRY=604800
```

4. ใช้ HTTPS เสมอ

5. พิจารณาใช้ key rotation strategy
