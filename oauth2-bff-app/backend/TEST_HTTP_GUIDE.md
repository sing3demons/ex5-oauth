# Backend API Testing with test.http

## Overview

The `test.http` file provides a comprehensive collection of HTTP requests for testing the backend API using the REST Client extension in VS Code.

## Prerequisites

### 1. Install REST Client Extension

Install the REST Client extension in VS Code:
- Open VS Code
- Go to Extensions (Ctrl+Shift+X / Cmd+Shift+X)
- Search for "REST Client" by Huachao Mao
- Click Install

Or install via command line:
```bash
code --install-extension humao.rest-client
```

### 2. Start Required Services

```bash
# Terminal 1: Start OAuth2 Server
cd /path/to/oauth2-server
./oauth2-server

# Terminal 2: Start MongoDB
mongod

# Terminal 3: Start Backend
cd oauth2-bff-app/backend
npm run dev
```

## How to Use

### Basic Usage

1. Open `test.http` file in VS Code
2. You'll see "Send Request" links above each `###` section
3. Click "Send Request" to execute the request
4. View response in a new panel

### Getting Access Token

Before testing protected endpoints, you need an access token:

#### Method 1: Through Browser (Recommended)

1. **Initiate Login**:
   ```http
   GET http://localhost:4000/auth/login
   ```
   Click "Send Request" and copy the `authorization_url`

2. **Complete OAuth2 Flow**:
   - Paste the URL in your browser
   - Login with your credentials
   - You'll be redirected to frontend with `access_token` in URL
   - Copy the `access_token` value

3. **Update Variable**:
   ```http
   @accessToken = paste_your_token_here
   ```

#### Method 2: Using Existing Session

If you're already logged in via the frontend:
1. Open browser DevTools (F12)
2. Go to Application > Cookies
3. Copy the `refresh_token` value
4. Use the refresh endpoint:
   ```http
   POST http://localhost:4000/auth/refresh
   Cookie: refresh_token=YOUR_REFRESH_TOKEN
   ```

### Testing Todos API

Once you have an access token:

1. **Get All Todos**:
   ```http
   GET http://localhost:4000/api/todos
   Authorization: Bearer {{accessToken}}
   ```

2. **Create Todo**:
   ```http
   POST http://localhost:4000/api/todos
   Authorization: Bearer {{accessToken}}
   Content-Type: application/json

   {
     "title": "My Todo",
     "description": "Description here",
     "status": "todo",
     "priority": "medium"
   }
   ```

3. **Copy Todo ID** from response:
   ```json
   {
     "id": "507f1f77bcf86cd799439011",  // Copy this
     ...
   }
   ```

4. **Update Variable**:
   ```http
   @todoId = 507f1f77bcf86cd799439011
   ```

5. **Update/Delete Todo**:
   Now you can use `{{todoId}}` in other requests

## Request Collections

### 1. Health & Discovery
- Health check
- OIDC discovery document
- JWKS endpoint

### 2. Authentication
- Initiate login
- OAuth2 callback
- Get user info
- Refresh token
- Logout
- Token validation

### 3. Todos CRUD
- Get all todos
- Create todo
- Update todo
- Update status (drag & drop)
- Delete todo

### 4. Error Cases
- Unauthorized requests
- Invalid tokens
- Validation errors
- Invalid IDs

### 5. Complete Flow
- Step-by-step example from login to logout

### 6. Bulk Operations
- Create multiple todos for testing

## Variables

The file uses variables for easy configuration:

```http
@baseUrl = http://localhost:4000
@oauth2Server = http://localhost:8080
@accessToken = YOUR_ACCESS_TOKEN_HERE
@refreshToken = YOUR_REFRESH_TOKEN_HERE
@todoId = YOUR_TODO_ID_HERE
```

Update these values as needed.

## Advanced Features

### Named Requests

Use `@name` to save responses:

```http
# @name createTodo
POST {{baseUrl}}/api/todos
...
```

Then reference the response:
```http
@todoId = {{createTodo.response.body.id}}
```

### Dynamic Values

REST Client provides built-in variables:

```http
{
  "title": "Todo {{$randomInt}}",
  "createdAt": "{{$timestamp}}",
  "id": "{{$guid}}"
}
```

### Environment Variables

Create `.vscode/settings.json`:

```json
{
  "rest-client.environmentVariables": {
    "local": {
      "baseUrl": "http://localhost:4000",
      "accessToken": "your_token_here"
    },
    "production": {
      "baseUrl": "https://api.example.com",
      "accessToken": "prod_token_here"
    }
  }
}
```

Switch environments in VS Code status bar.

## Common Workflows

### Workflow 1: Test Complete Auth Flow

1. Get authorization URL
2. Complete OAuth2 in browser
3. Get access token from callback
4. Test user info endpoint
5. Test refresh token
6. Test logout

### Workflow 2: Test Todos CRUD

1. Get access token
2. Create multiple todos
3. Get all todos
4. Update a todo
5. Change status (drag & drop)
6. Delete a todo

### Workflow 3: Test Error Handling

1. Try requests without token
2. Try with invalid token
3. Try with invalid data
4. Try with invalid IDs

## Tips & Tricks

### 1. Quick Testing

Use keyboard shortcuts:
- `Ctrl+Alt+R` / `Cmd+Alt+R`: Send request
- `Ctrl+Alt+C` / `Cmd+Alt+C`: Cancel request
- `Ctrl+Alt+E` / `Cmd+Alt+E`: Switch environment

### 2. Response Inspection

- Click on response to view in editor
- Use "Preview" tab for formatted JSON
- Use "Headers" tab to see response headers
- Use "Cookies" tab to see set cookies

### 3. Save Responses

Right-click on response and select "Save Response" to save to file.

### 4. Request History

View request history in REST Client panel (Ctrl+Alt+H / Cmd+Alt+H).

### 5. Code Generation

Right-click on request and select "Generate Code Snippet" to get code in various languages.

## Troubleshooting

### Issue: "Connection Refused"

**Solution**: Ensure backend server is running on port 4000
```bash
cd oauth2-bff-app/backend
npm run dev
```

### Issue: "401 Unauthorized"

**Solution**: 
1. Check if access token is valid
2. Get a new token using the login flow
3. Update `@accessToken` variable

### Issue: "CSRF Token Missing"

**Solution**: 
1. Get CSRF token first:
   ```http
   GET http://localhost:4000/csrf-token
   ```
2. Add to request headers:
   ```http
   X-CSRF-Token: token_value
   ```

### Issue: "Invalid Todo ID"

**Solution**:
1. Get todos to see valid IDs
2. Copy a valid ID
3. Update `@todoId` variable

### Issue: "MongoDB Connection Error"

**Solution**: Ensure MongoDB is running
```bash
mongod
# or
brew services start mongodb-community
```

## Best Practices

### 1. Use Variables

Always use variables for:
- Base URLs
- Tokens
- IDs
- Common values

### 2. Organize Requests

Group related requests together:
- Authentication
- CRUD operations
- Error cases

### 3. Add Comments

Document what each request does:
```http
### Create High Priority Todo
# This creates a todo with high priority
# Used for testing priority filtering
POST {{baseUrl}}/api/todos
...
```

### 4. Test Error Cases

Always test:
- Missing required fields
- Invalid data types
- Invalid IDs
- Unauthorized access

### 5. Clean Up

Delete test data after testing:
```http
### Cleanup - Delete Test Todos
DELETE {{baseUrl}}/api/todos/{{todoId}}
Authorization: Bearer {{accessToken}}
```

## Integration with CI/CD

You can run REST Client tests in CI/CD using the CLI:

```bash
# Install REST Client CLI
npm install -g rest-client-cli

# Run tests
rest-client test.http --env local
```

## Alternative Tools

If you prefer other tools:

### Postman
- Import `test.http` using Postman's import feature
- Or use the Postman collection generator

### cURL
- Right-click on request → "Copy Request As cURL"
- Paste in terminal

### HTTPie
- Right-click on request → "Copy Request As HTTPie"
- Paste in terminal

## Resources

- [REST Client Documentation](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)
- [REST Client GitHub](https://github.com/Huachao/vscode-restclient)
- [HTTP Syntax Reference](https://httpwg.org/specs/)

## Support

For issues or questions:
1. Check this guide
2. Review the `test.http` file comments
3. Check backend logs
4. Review API documentation

## Summary

The `test.http` file provides:
- ✅ Complete API coverage
- ✅ Authentication flow testing
- ✅ CRUD operations testing
- ✅ Error case testing
- ✅ Easy to use with REST Client
- ✅ Well documented with examples
- ✅ Variables for easy configuration

Happy testing! 🚀
