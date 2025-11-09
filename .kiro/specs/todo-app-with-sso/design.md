# Todo App with OAuth2/OIDC SSO - Design Document

## Overview

A modern, responsive Todo application with OAuth2/OIDC Single Sign-On integration and drag-and-drop functionality. The application consists of a React frontend and a Node.js/Express backend, integrating with the existing OAuth2 server for authentication.

## Architecture

### High-Level Architecture

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│                 │         │                  │         │                 │
│  React Frontend │◄───────►│  Express Backend │◄───────►│  OAuth2 Server  │
│  (Port 3000)    │         │  (Port 4000)     │         │  (Port 8080)    │
│                 │         │                  │         │                 │
└─────────────────┘         └──────────────────┘         └─────────────────┘
                                     │
                                     ▼
                            ┌──────────────────┐
                            │                  │
                            │  MongoDB         │
                            │  (Port 27017)    │
                            │                  │
                            └──────────────────┘
```

### Technology Stack

**Frontend:**
- React 18+ with TypeScript
- React DnD (Drag and Drop)
- Axios for HTTP requests
- React Router for navigation
- Tailwind CSS for styling
- React Query for data fetching and caching

**Backend:**
- Node.js with Express
- TypeScript
- MongoDB with Mongoose
- Passport.js for OAuth2 authentication
- Express Session for session management
- Helmet for security headers
- CORS for cross-origin requests

**OAuth2 Integration:**
- Existing OAuth2/OIDC Server (Go)
- Authorization Code Flow with PKCE
- JWT token validation
- SSO session management

## Components and Interfaces

### Frontend Components

#### 1. App Component
```typescript
interface AppProps {}

// Root component with routing and authentication context
```

#### 2. AuthProvider Component
```typescript
interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: () => void;
  logout: () => void;
  refreshToken: () => Promise<void>;
}

// Manages authentication state and token refresh
```

#### 3. LoginCallback Component
```typescript
interface LoginCallbackProps {}

// Handles OAuth2 callback and token exchange
```

#### 4. Dashboard Component
```typescript
interface DashboardProps {}

// Main todo board with drag-and-drop lists
```

#### 5. TodoBoard Component
```typescript
interface TodoBoardProps {
  todos: Todo[];
  onTodoMove: (todoId: string, newStatus: TodoStatus) => void;
  onTodoCreate: (todo: CreateTodoDto) => void;
  onTodoUpdate: (todoId: string, updates: UpdateTodoDto) => void;
  onTodoDelete: (todoId: string) => void;
}

// Container for all todo lists with drag-and-drop
```

#### 6. TodoList Component
```typescript
interface TodoListProps {
  title: string;
  status: TodoStatus;
  todos: Todo[];
  onDrop: (todoId: string) => void;
}

// Individual list column (To Do, In Progress, Done)
```

#### 7. TodoCard Component
```typescript
interface TodoCardProps {
  todo: Todo;
  onEdit: (todo: Todo) => void;
  onDelete: (todoId: string) => void;
  isDragging: boolean;
}

// Draggable todo item card
```

#### 8. TodoForm Component
```typescript
interface TodoFormProps {
  initialData?: Todo;
  onSubmit: (data: CreateTodoDto | UpdateTodoDto) => void;
  onCancel: () => void;
}

// Form for creating/editing todos
```

#### 9. Header Component
```typescript
interface HeaderProps {
  user: User;
  onLogout: () => void;
}

// Application header with user profile and logout
```

#### 10. ProtectedRoute Component
```typescript
interface ProtectedRouteProps {
  children: React.ReactNode;
}

// Route wrapper that requires authentication
```

### Backend API Endpoints

#### Authentication Endpoints

```typescript
// Initiate OAuth2 login
GET /auth/login
Response: Redirect to OAuth2 Server

// OAuth2 callback handler
GET /auth/callback?code=xxx&state=xxx
Response: { access_token, refresh_token, user }

// Refresh access token
POST /auth/refresh
Body: { refresh_token }
Response: { access_token }

// Logout
POST /auth/logout
Response: { message: "Logged out successfully" }

// Get current user
GET /auth/me
Headers: Authorization: Bearer {token}
Response: User
```

#### Todo Endpoints

```typescript
// Get all todos for authenticated user
GET /api/todos
Headers: Authorization: Bearer {token}
Response: Todo[]

// Create new todo
POST /api/todos
Headers: Authorization: Bearer {token}
Body: CreateTodoDto
Response: Todo

// Update todo
PATCH /api/todos/:id
Headers: Authorization: Bearer {token}
Body: UpdateTodoDto
Response: Todo

// Delete todo
DELETE /api/todos/:id
Headers: Authorization: Bearer {token}
Response: { message: "Todo deleted" }

// Move todo to different status
PATCH /api/todos/:id/move
Headers: Authorization: Bearer {token}
Body: { status: TodoStatus, position: number }
Response: Todo
```

## Data Models

### User Model (from OAuth2 Server)

```typescript
interface User {
  id: string;           // User ID from OAuth2 server (sub claim)
  email: string;        // User email
  name: string;         // User display name
  picture?: string;     // User profile picture URL
}
```

### Todo Model

```typescript
interface Todo {
  _id: string;                    // MongoDB ObjectId
  userId: string;                 // User ID from OAuth2 token
  title: string;                  // Todo title (max 200 chars)
  description?: string;           // Optional description (max 1000 chars)
  status: TodoStatus;             // Current status
  position: number;               // Position within the list
  createdAt: Date;                // Creation timestamp
  updatedAt: Date;                // Last update timestamp
}

enum TodoStatus {
  TODO = 'todo',
  IN_PROGRESS = 'in_progress',
  DONE = 'done'
}
```

### CreateTodoDto

```typescript
interface CreateTodoDto {
  title: string;                  // Required, 1-200 chars
  description?: string;           // Optional, max 1000 chars
  status?: TodoStatus;            // Optional, defaults to TODO
}
```

### UpdateTodoDto

```typescript
interface UpdateTodoDto {
  title?: string;                 // Optional, 1-200 chars
  description?: string;           // Optional, max 1000 chars
  status?: TodoStatus;            // Optional
  position?: number;              // Optional
}
```

### OAuth2 Token Response

```typescript
interface TokenResponse {
  access_token: string;           // JWT access token
  refresh_token: string;          // Refresh token
  token_type: 'Bearer';
  expires_in: number;             // Seconds until expiration
  id_token?: string;              // OIDC ID token
}
```

## OAuth2 Integration Flow

### 1. Initial Login Flow

```
User → Frontend → Backend → OAuth2 Server
  1. User clicks "Login"
  2. Frontend redirects to /auth/login
  3. Backend generates state and PKCE challenge
  4. Backend redirects to OAuth2 /oauth/authorize
  5. User authenticates on OAuth2 Server
  6. OAuth2 Server redirects to /auth/callback with code
  7. Backend exchanges code for tokens
  8. Backend validates tokens and extracts user info
  9. Backend creates session and returns tokens to frontend
  10. Frontend stores tokens and redirects to dashboard
```

### 2. SSO Flow (Returning User)

```
User → Frontend → Backend → OAuth2 Server
  1. User accesses app with valid SSO session
  2. Frontend redirects to /auth/login
  3. Backend redirects to OAuth2 /oauth/authorize
  4. OAuth2 Server validates SSO session (cookie)
  5. OAuth2 Server immediately returns authorization code
  6. Backend exchanges code for tokens
  7. Frontend receives tokens and shows dashboard
  
Total time: < 500ms (no user interaction needed)
```

### 3. Token Refresh Flow

```
Frontend → Backend → OAuth2 Server
  1. Access token expires (detected by 401 response)
  2. Frontend calls /auth/refresh with refresh_token
  3. Backend calls OAuth2 /oauth/token with refresh_token
  4. OAuth2 Server validates and issues new access_token
  5. Backend returns new access_token to frontend
  6. Frontend retries original request with new token
```

### 4. Logout Flow

```
User → Frontend → Backend → OAuth2 Server
  1. User clicks "Logout"
  2. Frontend calls /auth/logout
  3. Backend clears session
  4. Backend redirects to OAuth2 /auth/logout
  5. OAuth2 Server clears SSO session
  6. OAuth2 Server redirects back to app
  7. Frontend clears tokens and shows login page
```

## Drag-and-Drop Implementation

### React DnD Setup

```typescript
// DnD Context Provider
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { TouchBackend } from 'react-dnd-touch-backend';

// Use HTML5 backend for desktop, Touch backend for mobile
const backend = isMobile ? TouchBackend : HTML5Backend;

<DndProvider backend={backend}>
  <TodoBoard />
</DndProvider>
```

### Draggable Todo Card

```typescript
import { useDrag } from 'react-dnd';

const TodoCard: React.FC<TodoCardProps> = ({ todo }) => {
  const [{ isDragging }, drag] = useDrag({
    type: 'TODO',
    item: { id: todo._id, status: todo.status },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  });

  return (
    <div ref={drag} style={{ opacity: isDragging ? 0.5 : 1 }}>
      {/* Card content */}
    </div>
  );
};
```

### Droppable Todo List

```typescript
import { useDrop } from 'react-dnd';

const TodoList: React.FC<TodoListProps> = ({ status, onDrop }) => {
  const [{ isOver }, drop] = useDrop({
    accept: 'TODO',
    drop: (item: { id: string; status: TodoStatus }) => {
      if (item.status !== status) {
        onDrop(item.id, status);
      }
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
    }),
  });

  return (
    <div ref={drop} className={isOver ? 'bg-blue-100' : ''}>
      {/* List content */}
    </div>
  );
};
```

## Security Implementation

### 1. Token Storage

**Frontend:**
```typescript
// Store tokens in memory (not localStorage for security)
class TokenManager {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  setTokens(access: string, refresh: string) {
    this.accessToken = access;
    this.refreshToken = refresh;
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  clear() {
    this.accessToken = null;
    this.refreshToken = null;
  }
}
```

**Backend:**
```typescript
// Use HTTP-only cookies for refresh tokens
app.use(session({
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false,
  cookie: {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
  },
}));
```

### 2. Token Validation Middleware

```typescript
import jwt from 'jsonwebtoken';
import axios from 'axios';

// Middleware to validate access token
export const authenticateToken = async (req, res, next) => {
  const authHeader = req.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1];

  if (!token) {
    return res.status(401).json({ error: 'Access token required' });
  }

  try {
    // Verify token signature with OAuth2 server's public key
    const publicKey = await getPublicKey(); // Fetch from JWKS endpoint
    const decoded = jwt.verify(token, publicKey, { algorithms: ['RS256'] });
    
    req.user = {
      id: decoded.sub,
      email: decoded.email,
      name: decoded.name,
    };
    
    next();
  } catch (error) {
    return res.status(403).json({ error: 'Invalid token' });
  }
};
```

### 3. CSRF Protection

```typescript
import csrf from 'csurf';

// CSRF protection for state-changing operations
const csrfProtection = csrf({ cookie: true });

app.post('/api/todos', csrfProtection, authenticateToken, createTodo);
app.patch('/api/todos/:id', csrfProtection, authenticateToken, updateTodo);
app.delete('/api/todos/:id', csrfProtection, authenticateToken, deleteTodo);
```

### 4. CORS Configuration

```typescript
import cors from 'cors';

app.use(cors({
  origin: process.env.FRONTEND_URL || 'http://localhost:3000',
  credentials: true,
  methods: ['GET', 'POST', 'PATCH', 'DELETE'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-CSRF-Token'],
}));
```

## Error Handling

### Frontend Error Handling

```typescript
// Axios interceptor for automatic token refresh
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // If 401 and not already retried, try to refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const { data } = await axios.post('/auth/refresh');
        tokenManager.setTokens(data.access_token, data.refresh_token);
        
        // Retry original request with new token
        originalRequest.headers['Authorization'] = `Bearer ${data.access_token}`;
        return axios(originalRequest);
      } catch (refreshError) {
        // Refresh failed, redirect to login
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
```

### Backend Error Handling

```typescript
// Global error handler
app.use((err, req, res, next) => {
  console.error('Error:', err);

  // OAuth2 errors
  if (err.name === 'OAuthError') {
    return res.status(401).json({
      error: 'authentication_failed',
      message: 'OAuth2 authentication failed',
    });
  }

  // Validation errors
  if (err.name === 'ValidationError') {
    return res.status(400).json({
      error: 'validation_error',
      message: err.message,
    });
  }

  // Database errors
  if (err.name === 'MongoError') {
    return res.status(500).json({
      error: 'database_error',
      message: 'Database operation failed',
    });
  }

  // Default error
  res.status(500).json({
    error: 'internal_error',
    message: 'An unexpected error occurred',
  });
});
```

## Testing Strategy

### Frontend Testing

**Unit Tests:**
- Component rendering tests (React Testing Library)
- Hook tests (custom hooks for auth, todos)
- Utility function tests

**Integration Tests:**
- OAuth2 flow simulation
- Drag-and-drop interactions
- Form submissions
- API integration tests with MSW (Mock Service Worker)

**E2E Tests:**
- Complete user flows (Cypress or Playwright)
- Login → Create Todo → Drag → Logout

### Backend Testing

**Unit Tests:**
- Controller logic tests
- Middleware tests (authentication, validation)
- Service layer tests

**Integration Tests:**
- API endpoint tests with supertest
- Database operations with test database
- OAuth2 integration tests with mock server

**Security Tests:**
- Token validation tests
- CSRF protection tests
- Authorization tests (user can only access own todos)

### Test Coverage Goals

- Frontend: > 80% coverage
- Backend: > 85% coverage
- Critical paths: 100% coverage (auth, data persistence)

## Performance Optimization

### Frontend Optimization

1. **Code Splitting:**
```typescript
// Lazy load routes
const Dashboard = lazy(() => import('./pages/Dashboard'));
const LoginCallback = lazy(() => import('./pages/LoginCallback'));
```

2. **React Query Caching:**
```typescript
// Cache todo data with automatic refetching
const { data: todos } = useQuery('todos', fetchTodos, {
  staleTime: 30000, // 30 seconds
  cacheTime: 300000, // 5 minutes
});
```

3. **Optimistic Updates:**
```typescript
// Update UI immediately, rollback on error
const mutation = useMutation(updateTodo, {
  onMutate: async (newTodo) => {
    await queryClient.cancelQueries('todos');
    const previousTodos = queryClient.getQueryData('todos');
    queryClient.setQueryData('todos', (old) => [...old, newTodo]);
    return { previousTodos };
  },
  onError: (err, newTodo, context) => {
    queryClient.setQueryData('todos', context.previousTodos);
  },
});
```

### Backend Optimization

1. **Database Indexing:**
```typescript
// MongoDB indexes for fast queries
todoSchema.index({ userId: 1, status: 1 });
todoSchema.index({ userId: 1, position: 1 });
```

2. **Response Caching:**
```typescript
// Cache user info from OAuth2 server
const userCache = new Map();

async function getUserInfo(accessToken: string) {
  if (userCache.has(accessToken)) {
    return userCache.get(accessToken);
  }
  
  const userInfo = await fetchUserInfo(accessToken);
  userCache.set(accessToken, userInfo);
  
  // Clear cache after 5 minutes
  setTimeout(() => userCache.delete(accessToken), 300000);
  
  return userInfo;
}
```

3. **Connection Pooling:**
```typescript
// MongoDB connection pool
mongoose.connect(process.env.MONGODB_URI, {
  maxPoolSize: 10,
  minPoolSize: 2,
});
```

## Deployment Architecture

### Development Environment

```
Frontend:  http://localhost:3000
Backend:   http://localhost:4000
OAuth2:    http://localhost:8080
MongoDB:   mongodb://localhost:27017
```

### Production Environment

```
Frontend:  https://todo.example.com
Backend:   https://api.todo.example.com
OAuth2:    https://auth.example.com
MongoDB:   MongoDB Atlas or self-hosted
```

### Environment Variables

**Frontend (.env):**
```bash
REACT_APP_API_URL=http://localhost:4000
REACT_APP_OAUTH2_URL=http://localhost:8080
```

**Backend (.env):**
```bash
# Server
PORT=4000
NODE_ENV=development

# OAuth2
OAUTH2_SERVER_URL=http://localhost:8080
OAUTH2_CLIENT_ID=todo-app
OAUTH2_CLIENT_SECRET=your-client-secret
OAUTH2_REDIRECT_URI=http://localhost:4000/auth/callback

# Database
MONGODB_URI=mongodb://localhost:27017/todo_app

# Security
SESSION_SECRET=your-session-secret
FRONTEND_URL=http://localhost:3000

# CORS
CORS_ORIGIN=http://localhost:3000
```

## Database Schema

### MongoDB Collections

#### todos Collection

```javascript
{
  _id: ObjectId("..."),
  userId: "user-id-from-oauth2",
  title: "Complete project documentation",
  description: "Write comprehensive docs for the todo app",
  status: "in_progress",
  position: 0,
  createdAt: ISODate("2025-11-09T10:00:00Z"),
  updatedAt: ISODate("2025-11-09T15:30:00Z")
}
```

**Indexes:**
```javascript
db.todos.createIndex({ userId: 1, status: 1 });
db.todos.createIndex({ userId: 1, position: 1 });
db.todos.createIndex({ createdAt: -1 });
```

## UI/UX Design

### Color Scheme

```css
:root {
  --primary: #3b82f6;      /* Blue */
  --secondary: #8b5cf6;    /* Purple */
  --success: #10b981;      /* Green */
  --warning: #f59e0b;      /* Orange */
  --danger: #ef4444;       /* Red */
  --background: #f9fafb;   /* Light gray */
  --surface: #ffffff;      /* White */
  --text: #111827;         /* Dark gray */
  --text-secondary: #6b7280; /* Medium gray */
}
```

### Layout Structure

```
┌─────────────────────────────────────────────────────┐
│  Header (User Profile, Logout)                      │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │ To Do    │  │ In Prog  │  │ Done     │         │
│  │ (3)      │  │ (2)      │  │ (5)      │         │
│  ├──────────┤  ├──────────┤  ├──────────┤         │
│  │ [+] Add  │  │          │  │          │         │
│  ├──────────┤  ├──────────┤  ├──────────┤         │
│  │ □ Task 1 │  │ □ Task 4 │  │ ☑ Task 7 │         │
│  │ □ Task 2 │  │ □ Task 5 │  │ ☑ Task 8 │         │
│  │ □ Task 3 │  │          │  │ ☑ Task 9 │         │
│  │          │  │          │  │ ☑ Task10 │         │
│  │          │  │          │  │ ☑ Task11 │         │
│  └──────────┘  └──────────┘  └──────────┘         │
│                                                      │
└─────────────────────────────────────────────────────┘
```

### Responsive Breakpoints

```css
/* Mobile: < 768px - Stacked lists */
/* Tablet: 768px - 1024px - 2 columns */
/* Desktop: > 1024px - 3 columns */
```

## Future Enhancements

1. **Collaborative Features:**
   - Share todos with other users
   - Real-time updates with WebSockets
   - Comments on todos

2. **Advanced Features:**
   - Due dates and reminders
   - Priority levels
   - Tags and categories
   - Search and filtering
   - Recurring todos

3. **Customization:**
   - Custom lists (beyond To Do, In Progress, Done)
   - Themes and color schemes
   - Keyboard shortcuts

4. **Analytics:**
   - Productivity metrics
   - Completion rates
   - Time tracking

5. **Mobile App:**
   - React Native mobile app
   - Offline support with sync
   - Push notifications
