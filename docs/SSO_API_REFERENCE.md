# SSO API Reference

This document provides detailed API reference for the Single Sign-On (SSO) session management and authorization management endpoints.

## Table of Contents

- [Authentication](#authentication)
- [Session Management Endpoints](#session-management-endpoints)
- [Authorization Management Endpoints](#authorization-management-endpoints)
- [Error Responses](#error-responses)
- [Rate Limiting](#rate-limiting)

## Authentication

All session and authorization management endpoints require authentication via Bearer token in the Authorization header.

```http
Authorization: Bearer <access_token>
```

The access token must be a valid JWT token issued by the OAuth2 server. The user ID is extracted from the token's `sub` claim.

## Session Management Endpoints

### List Active Sessions

Retrieve all active SSO sessions for the authenticated user.

**Endpoint**: `GET /account/sessions`

**Authentication**: Required (Bearer token)

**Request**:
```http
GET /account/sessions HTTP/1.1
Host: localhost:8080
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**: `200 OK`

```json
{
  "sessions": [
    {
      "session_id": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
      "created_at": "2025-11-09T10:00:00Z",
      "last_activity": "2025-11-09T15:30:00Z",
      "expires_at": "2025-11-16T10:00:00Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
    },
    {
      "session_id": "q1r2s3t4u5v6w7x8y9z0a1b2c3d4e5f6",
      "created_at": "2025-11-08T14:00:00Z",
      "last_activity": "2025-11-09T09:00:00Z",
      "expires_at": "2025-11-15T14:00:00Z",
      "ip_address": "192.168.1.50",
      "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15"
    }
  ]
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `sessions` | Array | List of active SSO sessions |
| `sessions[].session_id` | String | Unique session identifier (32-byte random string) |
| `sessions[].created_at` | String (ISO 8601) | Session creation timestamp |
| `sessions[].last_activity` | String (ISO 8601) | Last activity timestamp |
| `sessions[].expires_at` | String (ISO 8601) | Session expiration timestamp (7 days from creation) |
| `sessions[].ip_address` | String | Client IP address at session creation |
| `sessions[].user_agent` | String | Browser/device user agent string |

**Error Responses**:

| Status Code | Error | Description |
|-------------|-------|-------------|
| `401 Unauthorized` | `invalid_token` | Missing or invalid access token |
| `500 Internal Server Error` | `server_error` | Database or internal error |

**Example (cURL)**:

```bash
curl -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Use Cases**:
- Security audit: View all active sessions
- Detect suspicious sessions (unknown IP/device)
- Monitor session activity across devices

---

### Revoke Specific Session

Revoke a specific SSO session by session ID. This effectively logs out the user from that specific device/browser.

**Endpoint**: `DELETE /account/sessions/{session_id}`

**Authentication**: Required (Bearer token)

**Path Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session_id` | String | Yes | The session ID to revoke |

**Request**:
```http
DELETE /account/sessions/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6 HTTP/1.1
Host: localhost:8080
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**: `200 OK`

```json
{
  "message": "Session revoked successfully"
}
```

**Error Responses**:

| Status Code | Error | Description |
|-------------|-------|-------------|
| `400 Bad Request` | `invalid_request` | Missing or invalid session_id parameter |
| `401 Unauthorized` | `invalid_token` | Missing or invalid access token |
| `403 Forbidden` | `forbidden` | Session belongs to a different user |
| `404 Not Found` | `not_found` | Session not found |
| `500 Internal Server Error` | `server_error` | Database or internal error |

**Example (cURL)**:

```bash
curl -X DELETE http://localhost:8080/account/sessions/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6 \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Use Cases**:
- Logout from a specific device
- Revoke suspicious session
- Remote logout (e.g., lost phone)

**Notes**:
- The session is immediately deleted from the database
- The SSO cookie on that device will become invalid
- The user will need to re-authenticate on that device
- Other sessions remain active

---

## Authorization Management Endpoints

### List Authorized Applications

Retrieve all applications that have been granted access to the authenticated user's account.

**Endpoint**: `GET /account/authorizations`

**Authentication**: Required (Bearer token)

**Request**:
```http
GET /account/authorizations HTTP/1.1
Host: localhost:8080
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**: `200 OK`

```json
{
  "authorizations": [
    {
      "client_id": "my-app-a",
      "client_name": "My Application A",
      "scopes": ["openid", "profile", "email"],
      "granted_at": "2025-11-01T10:00:00Z",
      "expires_at": "2026-11-01T10:00:00Z"
    },
    {
      "client_id": "my-app-b",
      "client_name": "My Application B",
      "scopes": ["openid", "profile"],
      "granted_at": "2025-11-05T14:30:00Z",
      "expires_at": "2026-11-05T14:30:00Z"
    }
  ]
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `authorizations` | Array | List of authorized applications |
| `authorizations[].client_id` | String | Unique client identifier |
| `authorizations[].client_name` | String | Human-readable application name |
| `authorizations[].scopes` | Array of Strings | Granted OAuth2/OIDC scopes |
| `authorizations[].granted_at` | String (ISO 8601) | Consent grant timestamp |
| `authorizations[].expires_at` | String (ISO 8601) | Consent expiration timestamp (1 year from grant) |

**Common Scopes**:

| Scope | Description |
|-------|-------------|
| `openid` | OpenID Connect authentication |
| `profile` | Access to user profile information (name, etc.) |
| `email` | Access to user email address |
| `phone` | Access to user phone number |
| `address` | Access to user address information |
| `offline_access` | Request refresh token for offline access |

**Error Responses**:

| Status Code | Error | Description |
|-------------|-------|-------------|
| `401 Unauthorized` | `invalid_token` | Missing or invalid access token |
| `500 Internal Server Error` | `server_error` | Database or internal error |

**Example (cURL)**:

```bash
curl -X GET http://localhost:8080/account/authorizations \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Use Cases**:
- Review which apps have access to user data
- Audit application permissions
- Identify unused applications
- Privacy management

---

### Revoke Application Authorization

Revoke consent for a specific application. This removes the user's authorization for the application to access their data.

**Endpoint**: `DELETE /account/authorizations/{client_id}`

**Authentication**: Required (Bearer token)

**Path Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `client_id` | String | Yes | The client ID to revoke authorization for |

**Request**:
```http
DELETE /account/authorizations/my-app-b HTTP/1.1
Host: localhost:8080
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**: `200 OK`

```json
{
  "message": "Authorization revoked successfully"
}
```

**Error Responses**:

| Status Code | Error | Description |
|-------------|-------|-------------|
| `400 Bad Request` | `invalid_request` | Missing or invalid client_id parameter |
| `401 Unauthorized` | `invalid_token` | Missing or invalid access token |
| `403 Forbidden` | `forbidden` | Authorization belongs to a different user |
| `404 Not Found` | `not_found` | Authorization not found |
| `500 Internal Server Error` | `server_error` | Database or internal error |

**Example (cURL)**:

```bash
curl -X DELETE http://localhost:8080/account/authorizations/my-app-b \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Use Cases**:
- Remove access for unused applications
- Revoke permissions after security incident
- Privacy management
- Compliance with data access requests

**Important Notes**:
- The consent record is immediately deleted from the database
- The next authorization request will show the consent screen
- **Existing access tokens remain valid** until they expire
- **Existing refresh tokens remain valid** until they expire or are revoked separately
- To fully revoke access, the application should also revoke tokens via the token revocation endpoint (if implemented)

**Behavior After Revocation**:

1. User visits the application
2. Application initiates OAuth2 authorization flow
3. User has valid SSO session (no login required)
4. **Consent screen is shown** (consent was revoked)
5. User must re-approve or deny access

---

## Error Responses

All endpoints follow a consistent error response format:

```json
{
  "error": "error_code",
  "error_description": "Human-readable error description"
}
```

### Common Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `invalid_request` | 400 | Malformed request or missing required parameters |
| `invalid_token` | 401 | Missing, expired, or invalid access token |
| `forbidden` | 403 | User does not have permission to access the resource |
| `not_found` | 404 | Requested resource not found |
| `server_error` | 500 | Internal server error |

### Example Error Response

**Request**:
```http
DELETE /account/sessions/invalid_session_id HTTP/1.1
Host: localhost:8080
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**: `404 Not Found`

```json
{
  "error": "not_found",
  "error_description": "Session not found"
}
```

---

## Rate Limiting

**Current Implementation**: No rate limiting is currently enforced on these endpoints.

**Recommended Limits** (for future implementation):

| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| `GET /account/sessions` | 60 requests | per minute |
| `DELETE /account/sessions/{session_id}` | 10 requests | per minute |
| `GET /account/authorizations` | 60 requests | per minute |
| `DELETE /account/authorizations/{client_id}` | 10 requests | per minute |

**Rate Limit Headers** (future):

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1699545600
```

---

## Complete Workflow Examples

### Example 1: Security Audit

A user wants to review all active sessions and revoke suspicious ones.

```bash
#!/bin/bash

ACCESS_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# Step 1: List all active sessions
echo "Fetching active sessions..."
curl -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  | jq .

# Output:
# {
#   "sessions": [
#     {
#       "session_id": "abc123...",
#       "ip_address": "192.168.1.1",
#       "user_agent": "Chrome on Mac",
#       ...
#     },
#     {
#       "session_id": "xyz789...",
#       "ip_address": "203.0.113.42",  ← Suspicious IP
#       "user_agent": "Unknown Browser",
#       ...
#     }
#   ]
# }

# Step 2: Revoke suspicious session
echo "Revoking suspicious session..."
curl -X DELETE http://localhost:8080/account/sessions/xyz789... \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Output:
# {
#   "message": "Session revoked successfully"
# }

# Step 3: Verify session removed
echo "Verifying session removed..."
curl -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  | jq .

# Output: Only one session remains
```

### Example 2: Privacy Management

A user wants to review and revoke access for unused applications.

```bash
#!/bin/bash

ACCESS_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# Step 1: List all authorized applications
echo "Fetching authorized applications..."
curl -X GET http://localhost:8080/account/authorizations \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  | jq .

# Output:
# {
#   "authorizations": [
#     {
#       "client_id": "active-app",
#       "client_name": "Active Application",
#       "scopes": ["openid", "profile"],
#       "granted_at": "2025-11-09T10:00:00Z"
#     },
#     {
#       "client_id": "old-app",
#       "client_name": "Old Application",
#       "scopes": ["openid", "profile", "email"],
#       "granted_at": "2024-05-01T10:00:00Z"  ← Old consent
#     }
#   ]
# }

# Step 2: Revoke authorization for old app
echo "Revoking authorization for old-app..."
curl -X DELETE http://localhost:8080/account/authorizations/old-app \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Output:
# {
#   "message": "Authorization revoked successfully"
# }

# Step 3: Verify authorization removed
echo "Verifying authorization removed..."
curl -X GET http://localhost:8080/account/authorizations \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  | jq .

# Output: Only active-app remains
```

### Example 3: Logout from All Devices

A user wants to logout from all devices (revoke all sessions).

```bash
#!/bin/bash

ACCESS_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# Step 1: Get all session IDs
SESSION_IDS=$(curl -s -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  | jq -r '.sessions[].session_id')

# Step 2: Revoke each session
for SESSION_ID in $SESSION_IDS; do
  echo "Revoking session: $SESSION_ID"
  curl -X DELETE http://localhost:8080/account/sessions/$SESSION_ID \
    -H "Authorization: Bearer $ACCESS_TOKEN"
done

echo "All sessions revoked successfully"
```

---

## Integration Examples

### JavaScript/TypeScript (Fetch API)

```typescript
// Session Management Service
class SessionManagementService {
  private baseURL = 'http://localhost:8080';
  private accessToken: string;

  constructor(accessToken: string) {
    this.accessToken = accessToken;
  }

  private getHeaders(): HeadersInit {
    return {
      'Authorization': `Bearer ${this.accessToken}`,
      'Content-Type': 'application/json'
    };
  }

  // List all active sessions
  async listSessions(): Promise<Session[]> {
    const response = await fetch(`${this.baseURL}/account/sessions`, {
      method: 'GET',
      headers: this.getHeaders()
    });

    if (!response.ok) {
      throw new Error(`Failed to list sessions: ${response.statusText}`);
    }

    const data = await response.json();
    return data.sessions;
  }

  // Revoke specific session
  async revokeSession(sessionId: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/account/sessions/${sessionId}`, {
      method: 'DELETE',
      headers: this.getHeaders()
    });

    if (!response.ok) {
      throw new Error(`Failed to revoke session: ${response.statusText}`);
    }
  }

  // List authorized applications
  async listAuthorizations(): Promise<Authorization[]> {
    const response = await fetch(`${this.baseURL}/account/authorizations`, {
      method: 'GET',
      headers: this.getHeaders()
    });

    if (!response.ok) {
      throw new Error(`Failed to list authorizations: ${response.statusText}`);
    }

    const data = await response.json();
    return data.authorizations;
  }

  // Revoke application authorization
  async revokeAuthorization(clientId: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/account/authorizations/${clientId}`, {
      method: 'DELETE',
      headers: this.getHeaders()
    });

    if (!response.ok) {
      throw new Error(`Failed to revoke authorization: ${response.statusText}`);
    }
  }
}

// Usage
const service = new SessionManagementService(accessToken);

// List sessions
const sessions = await service.listSessions();
console.log('Active sessions:', sessions);

// Revoke session
await service.revokeSession('abc123...');
console.log('Session revoked');

// List authorizations
const authorizations = await service.listAuthorizations();
console.log('Authorized apps:', authorizations);

// Revoke authorization
await service.revokeAuthorization('old-app');
console.log('Authorization revoked');
```

### Python (requests library)

```python
import requests
from typing import List, Dict

class SessionManagementClient:
    def __init__(self, base_url: str, access_token: str):
        self.base_url = base_url
        self.access_token = access_token
        self.headers = {
            'Authorization': f'Bearer {access_token}',
            'Content-Type': 'application/json'
        }

    def list_sessions(self) -> List[Dict]:
        """List all active SSO sessions"""
        response = requests.get(
            f'{self.base_url}/account/sessions',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()['sessions']

    def revoke_session(self, session_id: str) -> None:
        """Revoke a specific SSO session"""
        response = requests.delete(
            f'{self.base_url}/account/sessions/{session_id}',
            headers=self.headers
        )
        response.raise_for_status()

    def list_authorizations(self) -> List[Dict]:
        """List all authorized applications"""
        response = requests.get(
            f'{self.base_url}/account/authorizations',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()['authorizations']

    def revoke_authorization(self, client_id: str) -> None:
        """Revoke application authorization"""
        response = requests.delete(
            f'{self.base_url}/account/authorizations/{client_id}',
            headers=self.headers
        )
        response.raise_for_status()

# Usage
client = SessionManagementClient('http://localhost:8080', access_token)

# List sessions
sessions = client.list_sessions()
print(f'Active sessions: {len(sessions)}')

# Revoke suspicious session
client.revoke_session('abc123...')
print('Session revoked')

# List authorizations
authorizations = client.list_authorizations()
for auth in authorizations:
    print(f"App: {auth['client_name']}, Scopes: {auth['scopes']}")

# Revoke authorization
client.revoke_authorization('old-app')
print('Authorization revoked')
```

---

## Additional Resources

- [SSO Usage Guide](./SSO_USAGE.md) - Comprehensive SSO usage examples
- [OAuth2 RFC 6749](https://tools.ietf.org/html/rfc6749) - OAuth2 specification
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html) - OIDC specification
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519) - JSON Web Token specification
