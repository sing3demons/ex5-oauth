# Requirements Document

## Introduction

สร้าง Backend for Frontend (BFF) application ด้วย NestJS framework เพื่อทดแทน Express version ปัจจุบัน โดยมีฟีเจอร์ครบถ้วนเหมือนเดิม รวมถึง OAuth2/OIDC authentication, Todo CRUD operations, MongoDB integration และ security features ทั้งหมด

## Glossary

- **NestJS Backend**: Backend application ที่สร้างด้วย NestJS framework พร้อม TypeScript
- **OAuth2 Server**: External OAuth2/OIDC authorization server ที่ทำงานที่ http://localhost:8080
- **BFF Pattern**: Backend for Frontend pattern ที่จัดการ authentication และ API calls
- **Confidential Client**: OAuth2 client ที่ใช้ client_secret ในการ authenticate
- **OIDC**: OpenID Connect protocol สำหรับ authentication
- **Todo Service**: Service สำหรับจัดการ todo items ของ user
- **MongoDB**: NoSQL database สำหรับเก็บ todo data
- **HttpOnly Cookie**: Secure cookie ที่ไม่สามารถเข้าถึงจาก JavaScript
- **Refresh Token**: Token สำหรับขอ access token ใหม่โดยไม่ต้อง login ใหม่
- **ID Token**: JWT token ที่มี user claims จาก OIDC
- **Access Token**: Token สำหรับเข้าถึง protected resources

## Requirements

### Requirement 1

**User Story:** ในฐานะ developer ฉันต้องการ NestJS project structure ที่เป็นมาตรฐาน เพื่อให้ง่ายต่อการ maintain และ scale

#### Acceptance Criteria

1. WHEN initializing the project, THE NestJS Backend SHALL create a standard NestJS project structure with modules, controllers, services, and guards
2. WHEN organizing code, THE NestJS Backend SHALL separate concerns into auth module, todos module, database module, and shared utilities
3. WHEN configuring the application, THE NestJS Backend SHALL use ConfigModule for environment variable management
4. WHEN setting up TypeScript, THE NestJS Backend SHALL configure strict type checking and path aliases
5. THE NestJS Backend SHALL include package.json with all required dependencies including @nestjs/core, @nestjs/common, @nestjs/config, @nestjs/axios, mongodb, and cookie-parser

### Requirement 2

**User Story:** ในฐานะ user ฉันต้องการ login ผ่าน OAuth2/OIDC flow เพื่อเข้าใช้งาน application อย่างปลอดภัย

#### Acceptance Criteria

1. WHEN user requests login, THE NestJS Backend SHALL generate authorization URL with state, nonce, and OIDC parameters
2. WHEN OAuth2 Server redirects back with authorization code, THE NestJS Backend SHALL exchange code for tokens using client_secret
3. WHEN receiving tokens, THE NestJS Backend SHALL validate ID token signature, issuer, audience, expiration, and nonce
4. WHEN tokens are valid, THE NestJS Backend SHALL store refresh token in HttpOnly cookie with secure flags
5. WHEN authentication succeeds, THE NestJS Backend SHALL redirect to frontend with access token and ID token in URL parameters

### Requirement 3

**User Story:** ในฐานะ user ฉันต้องการ refresh access token อัตโนมัติ เพื่อไม่ต้อง login ใหม่บ่อยๆ

#### Acceptance Criteria

1. WHEN access token expires, THE NestJS Backend SHALL accept refresh token from HttpOnly cookie
2. WHEN refresh token is valid, THE NestJS Backend SHALL exchange it for new access token using client_secret
3. WHEN receiving new tokens, THE NestJS Backend SHALL update refresh token cookie if token rotation is enabled
4. WHEN refresh token is invalid or expired, THE NestJS Backend SHALL clear the cookie and return 401 error
5. THE NestJS Backend SHALL protect refresh endpoint with guard that validates cookie presence

### Requirement 4

**User Story:** ในฐานะ user ฉันต้องการ logout จาก application เพื่อยกเลิก session

#### Acceptance Criteria

1. WHEN user requests logout, THE NestJS Backend SHALL clear refresh token cookie with proper flags
2. WHEN logout succeeds, THE NestJS Backend SHALL return success response
3. THE NestJS Backend SHALL set cookie with httpOnly, secure, and sameSite flags when clearing

### Requirement 5

**User Story:** ในฐานะ user ฉันต้องการดึงข้อมูล user profile จาก OAuth2 server เพื่อแสดงในหน้า dashboard

#### Acceptance Criteria

1. WHEN user requests userinfo, THE NestJS Backend SHALL require valid access token in Authorization header
2. WHEN access token is present, THE NestJS Backend SHALL forward request to OAuth2 Server userinfo endpoint
3. WHEN OAuth2 Server responds, THE NestJS Backend SHALL return user claims to frontend
4. WHEN access token is missing or invalid, THE NestJS Backend SHALL return 401 error

### Requirement 6

**User Story:** ในฐานะ user ฉันต้องการสร้าง todo items เพื่อจัดการงานของฉัน

#### Acceptance Criteria

1. WHEN user creates todo, THE NestJS Backend SHALL require valid access token
2. WHEN todo data is valid, THE NestJS Backend SHALL extract user ID from token claims
3. WHEN user ID is extracted, THE NestJS Backend SHALL generate unique ID for todo using UUID
4. WHEN saving todo, THE NestJS Backend SHALL store todo in MongoDB with userId, title, description, status, priority, and timestamps
5. WHEN todo is created, THE NestJS Backend SHALL return created todo with 201 status code

### Requirement 7

**User Story:** ในฐานะ user ฉันต้องการดู todos ทั้งหมดของฉัน เพื่อติดตามงาน

#### Acceptance Criteria

1. WHEN user requests todos, THE NestJS Backend SHALL require valid access token
2. WHEN access token is valid, THE NestJS Backend SHALL extract user ID from token
3. WHEN querying database, THE NestJS Backend SHALL filter todos by userId
4. WHEN returning results, THE NestJS Backend SHALL sort todos by createdAt in descending order
5. THE NestJS Backend SHALL return only todos that belong to the authenticated user

### Requirement 8

**User Story:** ในฐานะ user ฉันต้องการแก้ไข todo items เพื่ออัพเดทข้อมูล

#### Acceptance Criteria

1. WHEN user updates todo, THE NestJS Backend SHALL require valid access token
2. WHEN todo exists, THE NestJS Backend SHALL verify that todo belongs to authenticated user
3. WHEN user is authorized, THE NestJS Backend SHALL update allowed fields including title, description, status, and priority
4. WHEN updating, THE NestJS Backend SHALL set updatedAt timestamp to current time
5. WHEN update succeeds, THE NestJS Backend SHALL return updated todo object

### Requirement 9

**User Story:** ในฐานะ user ฉันต้องการลบ todo items เพื่อเอางานที่เสร็จแล้วออก

#### Acceptance Criteria

1. WHEN user deletes todo, THE NestJS Backend SHALL require valid access token
2. WHEN todo exists, THE NestJS Backend SHALL verify that todo belongs to authenticated user
3. WHEN user is authorized, THE NestJS Backend SHALL remove todo from MongoDB
4. WHEN deletion succeeds, THE NestJS Backend SHALL return 204 No Content status
5. WHEN todo does not exist, THE NestJS Backend SHALL return 404 error

### Requirement 10

**User Story:** ในฐานะ user ฉันต้องการ drag & drop todos ระหว่าง columns เพื่อเปลี่ยน status

#### Acceptance Criteria

1. WHEN user changes todo status, THE NestJS Backend SHALL require valid access token
2. WHEN status value is provided, THE NestJS Backend SHALL validate that status is one of: todo, in_progress, or done
3. WHEN status is valid, THE NestJS Backend SHALL verify todo ownership
4. WHEN user is authorized, THE NestJS Backend SHALL update todo status and updatedAt timestamp
5. WHEN update succeeds, THE NestJS Backend SHALL return updated todo object

### Requirement 11

**User Story:** ในฐานะ developer ฉันต้องการ MongoDB integration เพื่อเก็บข้อมูลแบบ persistent

#### Acceptance Criteria

1. WHEN application starts, THE NestJS Backend SHALL connect to MongoDB using connection string from environment
2. WHEN connection succeeds, THE NestJS Backend SHALL create database instance accessible throughout application
3. WHEN storing todos, THE NestJS Backend SHALL use todos collection with proper indexes
4. WHEN connection fails, THE NestJS Backend SHALL log error and prevent application startup
5. THE NestJS Backend SHALL provide database service that can be injected into other services

### Requirement 12

**User Story:** ในฐานะ developer ฉันต้องการ CORS configuration เพื่อให้ frontend เรียก API ได้

#### Acceptance Criteria

1. WHEN receiving requests, THE NestJS Backend SHALL allow requests from configured frontend URL
2. WHEN handling CORS, THE NestJS Backend SHALL enable credentials for cookie support
3. WHEN configuring methods, THE NestJS Backend SHALL allow GET, POST, PUT, DELETE, PATCH, and OPTIONS methods
4. WHEN setting headers, THE NestJS Backend SHALL allow Content-Type and Authorization headers
5. THE NestJS Backend SHALL read frontend URL from environment configuration

### Requirement 13

**User Story:** ในฐานะ developer ฉันต้องการ authentication guards เพื่อป้องกัน protected endpoints

#### Acceptance Criteria

1. WHEN protecting endpoints, THE NestJS Backend SHALL provide AuthGuard that validates access token
2. WHEN validating token, THE NestJS Backend SHALL check Authorization header format
3. WHEN token is missing, THE NestJS Backend SHALL return 401 Unauthorized error
4. WHEN token is present, THE NestJS Backend SHALL allow request to proceed
5. THE NestJS Backend SHALL provide RefreshTokenGuard that validates refresh token cookie

### Requirement 14

**User Story:** ในฐานะ developer ฉันต้องการ OIDC utilities เพื่อจัดการ ID token validation

#### Acceptance Criteria

1. WHEN validating ID token, THE NestJS Backend SHALL decode JWT and extract claims
2. WHEN checking claims, THE NestJS Backend SHALL verify issuer matches OAuth2 server URL
3. WHEN checking audience, THE NestJS Backend SHALL verify aud claim matches client ID
4. WHEN checking expiration, THE NestJS Backend SHALL verify exp claim is in the future
5. WHEN nonce is provided, THE NestJS Backend SHALL verify nonce claim matches expected value

### Requirement 15

**User Story:** ในฐานะ developer ฉันต้องการ health check endpoint เพื่อตรวจสอบสถานะ application

#### Acceptance Criteria

1. WHEN health check is requested, THE NestJS Backend SHALL return status and timestamp
2. THE NestJS Backend SHALL respond with 200 OK status when application is healthy
3. THE NestJS Backend SHALL include current ISO timestamp in response
4. THE NestJS Backend SHALL not require authentication for health check endpoint
5. THE NestJS Backend SHALL expose health check at /health path

### Requirement 16

**User Story:** ในฐานะ developer ฉันต้องการ error handling middleware เพื่อจัดการ errors แบบ centralized

#### Acceptance Criteria

1. WHEN error occurs, THE NestJS Backend SHALL catch error using exception filter
2. WHEN logging error, THE NestJS Backend SHALL include error message and stack trace
3. WHEN returning error, THE NestJS Backend SHALL format response with error code and message
4. WHEN error has status code, THE NestJS Backend SHALL use that status code in response
5. WHEN error has no status code, THE NestJS Backend SHALL default to 500 Internal Server Error

### Requirement 17

**User Story:** ในฐานะ developer ฉันต้องการ session management เพื่อเก็บ state และ nonce ระหว่าง OAuth flow

#### Acceptance Criteria

1. WHEN initiating OAuth flow, THE NestJS Backend SHALL store state and nonce in memory with timestamp
2. WHEN callback is received, THE NestJS Backend SHALL retrieve session data using state parameter
3. WHEN session is found, THE NestJS Backend SHALL delete session after use
4. WHEN session is expired, THE NestJS Backend SHALL reject callback with invalid_state error
5. THE NestJS Backend SHALL clean up expired sessions periodically every 60 seconds

### Requirement 18

**User Story:** ในฐานะ developer ฉันต้องการ OIDC discovery endpoint เพื่อดึง configuration จาก OAuth2 server

#### Acceptance Criteria

1. WHEN discovery is requested, THE NestJS Backend SHALL fetch .well-known/openid-configuration from OAuth2 Server
2. WHEN OAuth2 Server responds, THE NestJS Backend SHALL return discovery document to client
3. WHEN request fails, THE NestJS Backend SHALL return 500 error with descriptive message
4. THE NestJS Backend SHALL not require authentication for discovery endpoint
5. THE NestJS Backend SHALL expose discovery at /auth/discovery path

### Requirement 19

**User Story:** ในฐานะ developer ฉันต้องการ token validation utilities เพื่อ validate และ decode tokens

#### Acceptance Criteria

1. WHEN validating token, THE NestJS Backend SHALL provide endpoint that accepts ID token
2. WHEN token is valid, THE NestJS Backend SHALL return validation result with claims
3. WHEN token is invalid, THE NestJS Backend SHALL return 401 error with reason
4. WHEN decoding token, THE NestJS Backend SHALL provide endpoint that decodes JWT without validation
5. THE NestJS Backend SHALL expose validation at /auth/validate-token and decode at /auth/decode-token

### Requirement 20

**User Story:** ในฐานะ developer ฉันต้องการ session info endpoint เพื่อ debug authentication state

#### Acceptance Criteria

1. WHEN session info is requested, THE NestJS Backend SHALL require refresh token cookie
2. WHEN cookie exists, THE NestJS Backend SHALL decode refresh token to extract claims
3. WHEN decoding succeeds, THE NestJS Backend SHALL return expiration time, user ID, and scope
4. WHEN decoding fails, THE NestJS Backend SHALL return error message
5. THE NestJS Backend SHALL expose session info at /auth/session path
