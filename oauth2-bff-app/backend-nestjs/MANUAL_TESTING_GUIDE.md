# Manual Testing Guide for NestJS Backend

This guide provides step-by-step instructions for manually testing the NestJS backend with the existing frontend.

## Prerequisites

Ensure the following services are running:

1. **OAuth2 Server** (Port 8080)
   ```bash
   ./oauth2-server
   ```

2. **MongoDB** (Port 27017)
   ```bash
   # Should already be running
   # Verify with: pgrep mongod
   ```

3. **NestJS Backend** (Port 3001)
   ```bash
   cd oauth2-bff-app/backend-nestjs
   npm run start:dev
   ```

4. **Frontend** (Port 5173)
   ```bash
   cd oauth2-bff-app/frontend
   npm run dev
   ```

## Test Scenarios

### 1. OAuth2 Login Flow

**Objective**: Verify that users can log in using OAuth2/OIDC

**Steps**:
1. Open browser to http://localhost:5173
2. You should see the login page
3. Click the "Login" button
4. You will be redirected to the OAuth2 server login page (http://localhost:8080)
5. Enter test credentials:
   - Username: `testuser` (or create a new user)
   - Password: `password`
6. Click "Login" or "Authorize"
7. You should be redirected back to the frontend dashboard

**Expected Results**:
- ✅ Redirect to OAuth2 server works
- ✅ After login, redirect back to frontend
- ✅ Dashboard displays user information
- ✅ Access token is stored in frontend
- ✅ Refresh token is stored in HttpOnly cookie (check DevTools > Application > Cookies)

**Verification**:
- Open DevTools (F12) > Network tab
- Check the `/auth/callback` request
- Verify the redirect includes `access_token` and `id_token` parameters
- Check Application > Cookies > http://localhost:3001
- Verify `refresh_token` cookie exists with HttpOnly flag

---

### 2. Token Refresh Flow

**Objective**: Verify that access tokens can be refreshed automatically

**Steps**:
1. Login to the application (follow Test 1)
2. Open DevTools > Network tab
3. Wait for the access token to expire (or manually trigger a refresh)
4. Make an API request (e.g., fetch todos)
5. Observe the network requests

**Expected Results**:
- ✅ When token expires, frontend calls `/auth/refresh`
- ✅ New access token is received
- ✅ Refresh token cookie is updated (if rotation enabled)
- ✅ User remains logged in
- ✅ API request succeeds with new token

**Verification**:
- Check Network tab for POST `/auth/refresh` request
- Verify response contains new `access_token`
- Verify subsequent requests use the new token

---

### 3. Todo CRUD Operations

**Objective**: Verify all todo operations work correctly

#### 3.1 Create Todo

**Steps**:
1. Login to the application
2. In the dashboard, find the "Create Todo" button or form
3. Enter todo details:
   - Title: "Test Todo"
   - Description: "This is a test"
   - Priority: "High"
4. Click "Create" or "Add"

**Expected Results**:
- ✅ Todo appears in the "To Do" column
- ✅ Todo has correct title, description, and priority
- ✅ Network request to POST `/api/todos` succeeds (201 status)

**Verification**:
- Open DevTools > Network tab
- Check POST `/api/todos` request
- Verify request body contains todo data
- Verify response contains created todo with `id`

#### 3.2 View Todos

**Steps**:
1. Login to the application
2. Dashboard should automatically load todos

**Expected Results**:
- ✅ All user's todos are displayed
- ✅ Todos are organized by status (To Do, In Progress, Done)
- ✅ Network request to GET `/api/todos` succeeds (200 status)

**Verification**:
- Check Network tab for GET `/api/todos` request
- Verify response contains array of todos
- Verify only current user's todos are returned

#### 3.3 Update Todo

**Steps**:
1. Login and view todos
2. Click on a todo to edit it
3. Modify the title, description, or priority
4. Save changes

**Expected Results**:
- ✅ Todo updates immediately in UI
- ✅ Network request to PUT `/api/todos/:id` succeeds (200 status)
- ✅ Changes persist after page refresh

**Verification**:
- Check Network tab for PUT `/api/todos/:id` request
- Verify request body contains updated data
- Verify response contains updated todo

#### 3.4 Delete Todo

**Steps**:
1. Login and view todos
2. Find the delete button on a todo
3. Click delete
4. Confirm deletion if prompted

**Expected Results**:
- ✅ Todo disappears from UI immediately
- ✅ Network request to DELETE `/api/todos/:id` succeeds (204 status)
- ✅ Todo does not reappear after page refresh

**Verification**:
- Check Network tab for DELETE `/api/todos/:id` request
- Verify response status is 204 No Content

---

### 4. Drag & Drop Status Updates

**Objective**: Verify that dragging todos between columns updates their status

**Steps**:
1. Login and create a few todos
2. Drag a todo from "To Do" column to "In Progress" column
3. Observe the UI and network requests
4. Drag the same todo from "In Progress" to "Done"
5. Refresh the page to verify persistence

**Expected Results**:
- ✅ Todo moves to new column immediately
- ✅ Network request to PATCH `/api/todos/:id/status` succeeds (200 status)
- ✅ Status persists after page refresh
- ✅ No errors in console

**Verification**:
- Open DevTools > Network tab
- Check PATCH `/api/todos/:id/status` request
- Verify request body: `{"status": "in_progress"}` or `{"status": "done"}`
- Verify response contains updated todo with new status

---

### 5. User Info Display

**Objective**: Verify that user information is fetched and displayed

**Steps**:
1. Login to the application
2. Look for user info display (usually in header or profile section)

**Expected Results**:
- ✅ User's name/email is displayed
- ✅ Network request to GET `/auth/userinfo` succeeds (200 status)
- ✅ User info matches OAuth2 server data

**Verification**:
- Check Network tab for GET `/auth/userinfo` request
- Verify Authorization header contains Bearer token
- Verify response contains user claims (sub, name, email, etc.)

---

### 6. Logout Flow

**Objective**: Verify that users can log out and session is cleared

**Steps**:
1. Login to the application
2. Find and click the "Logout" button
3. Observe the behavior

**Expected Results**:
- ✅ Network request to POST `/auth/logout` succeeds
- ✅ Refresh token cookie is cleared
- ✅ User is redirected to login page
- ✅ Cannot access protected routes without logging in again

**Verification**:
- Check Network tab for POST `/auth/logout` request
- Check Application > Cookies
- Verify `refresh_token` cookie is removed
- Try accessing `/api/todos` - should get 401 error

---

### 7. Error Handling

**Objective**: Verify that errors are handled gracefully

#### 7.1 Unauthorized Access

**Steps**:
1. Open browser in incognito/private mode
2. Try to access http://localhost:5173 directly
3. Try to manually call http://localhost:3001/api/todos

**Expected Results**:
- ✅ Frontend redirects to login page
- ✅ API returns 401 Unauthorized
- ✅ Error message is displayed appropriately

#### 7.2 Invalid Todo Operations

**Steps**:
1. Login to the application
2. Try to update a non-existent todo (manually call API)
3. Try to delete a non-existent todo

**Expected Results**:
- ✅ API returns 404 Not Found
- ✅ Error message is displayed
- ✅ Application doesn't crash

#### 7.3 Network Errors

**Steps**:
1. Login to the application
2. Stop the NestJS backend
3. Try to create or fetch todos

**Expected Results**:
- ✅ Frontend shows connection error
- ✅ User is informed of the issue
- ✅ Application remains functional when backend restarts

---

## Testing Checklist

Use this checklist to track your testing progress:

### OAuth2 Flow
- [ ] Login initiation works
- [ ] Redirect to OAuth2 server works
- [ ] Authorization works
- [ ] Callback redirect works
- [ ] Access token received
- [ ] Refresh token stored in cookie
- [ ] User info displayed

### Token Management
- [ ] Token refresh works automatically
- [ ] Refresh token rotation works (if enabled)
- [ ] Expired tokens handled correctly
- [ ] Invalid tokens rejected

### Todo Operations
- [ ] Create todo works
- [ ] View all todos works
- [ ] View single todo works
- [ ] Update todo works
- [ ] Delete todo works
- [ ] Status update works

### Drag & Drop
- [ ] Drag from "To Do" to "In Progress" works
- [ ] Drag from "In Progress" to "Done" works
- [ ] Drag from "Done" back to "In Progress" works
- [ ] Status persists after refresh

### Security
- [ ] Unauthorized requests rejected
- [ ] User can only access own todos
- [ ] Refresh token is HttpOnly
- [ ] CORS configured correctly
- [ ] Tokens validated properly

### Error Handling
- [ ] 401 errors handled
- [ ] 404 errors handled
- [ ] Network errors handled
- [ ] Invalid input handled
- [ ] Error messages displayed

### Logout
- [ ] Logout clears session
- [ ] Logout clears cookies
- [ ] Logout redirects to login
- [ ] Cannot access protected routes after logout

---

## Automated Testing

For automated testing without manual login, use the provided test scripts:

### Basic Integration Tests
```bash
cd oauth2-bff-app/backend-nestjs
./test-integration.sh
```

### Authenticated Tests (requires manual token)
```bash
# 1. Login via browser and get token from DevTools
# 2. Export token
export AUTH_TOKEN='Bearer your-token-here'

# 3. Run authenticated tests
./test-with-auth.sh
```

---

## Troubleshooting

### Issue: Cannot login
**Solution**: 
- Verify OAuth2 server is running on port 8080
- Check browser console for errors
- Verify client credentials in `.env` file

### Issue: Todos not loading
**Solution**:
- Verify NestJS backend is running on port 3001
- Check MongoDB is running
- Check browser console and network tab for errors
- Verify access token is valid

### Issue: Drag & drop not working
**Solution**:
- Check browser console for JavaScript errors
- Verify PATCH `/api/todos/:id/status` endpoint is working
- Check that status values are valid: 'todo', 'in_progress', 'done'

### Issue: Token refresh fails
**Solution**:
- Verify refresh token cookie exists
- Check cookie is not expired
- Verify OAuth2 server is running
- Check backend logs for errors

---

## Success Criteria

The NestJS backend is considered fully tested and working when:

1. ✅ All OAuth2 flows work correctly
2. ✅ All todo CRUD operations work
3. ✅ Drag & drop status updates work
4. ✅ Token refresh works automatically
5. ✅ Error handling works correctly
6. ✅ Security measures are in place
7. ✅ No console errors during normal operation
8. ✅ All automated tests pass
9. ✅ Frontend works seamlessly with backend
10. ✅ User experience is smooth and responsive

---

## Next Steps

After successful manual testing:

1. Document any issues found
2. Fix any bugs discovered
3. Perform load testing if needed
4. Deploy to staging environment
5. Conduct user acceptance testing
6. Deploy to production

---

## Support

If you encounter any issues during testing:

1. Check the backend logs in the terminal
2. Check browser console for frontend errors
3. Review the TESTING_RESULTS.md document
4. Check the MIGRATION.md for differences from Express version
5. Refer to the README.md for setup instructions
