# Implementation Plan

- [x] 1. Setup project structure and dependencies
  - Create monorepo structure with frontend and backend folders
  - Initialize React app with TypeScript and Tailwind CSS
  - Initialize Express backend with TypeScript
  - Install required dependencies (React DnD, Axios, Mongoose, Passport, etc.)
  - Setup environment configuration files
  - _Requirements: All requirements (foundation)_

- [x] 2. Register OAuth2 client with the OAuth2 server
  - Use the OAuth2 server's client registration endpoint
  - Configure redirect URI for the todo app backend
  - Store client credentials securely in environment variables
  - Document the client registration process
  - _Requirements: 1.1, 1.2_

- [x] 3. Implement backend OAuth2 authentication
- [x] 3.1 Create OAuth2 authentication routes
  - Implement /auth/login endpoint to initiate OAuth2 flow
  - Implement /auth/callback endpoint to handle authorization code
  - Generate and validate PKCE challenge and state parameters
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 3.2 Implement token exchange and validation
  - Exchange authorization code for access and refresh tokens
  - Validate JWT tokens with OAuth2 server's public key (JWKS)
  - Extract user information from token claims
  - _Requirements: 1.3, 1.4_

- [x] 3.3 Create session management
  - Setup Express session with secure cookies
  - Store refresh tokens in HTTP-only cookies
  - Implement /auth/refresh endpoint for token refresh
  - Implement /auth/logout endpoint
  - _Requirements: 1.5, 7.1, 7.2, 7.3_

- [x] 3.4 Create authentication middleware
  - Implement middleware to validate access tokens
  - Extract user ID from token for request context
  - Handle token expiration and refresh
  - _Requirements: 11.1, 11.2_

- [x] 3.5 Write authentication tests
  - Test OAuth2 flow with mock server
  - Test token validation and refresh
  - Test session management
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 4. Implement backend database and Todo API
- [x] 4.1 Setup MongoDB connection and Todo model
  - Configure Mongoose connection with connection pooling
  - Create Todo schema with validation
  - Add database indexes for performance
  - _Requirements: 9.1, 9.5_

- [x] 4.2 Implement Todo CRUD endpoints
  - Create GET /api/todos endpoint to fetch user's todos
  - Create POST /api/todos endpoint to create new todo
  - Create PATCH /api/todos/:id endpoint to update todo
  - Create DELETE /api/todos/:id endpoint to delete todo
  - Add authentication middleware to all endpoints
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 9.2, 9.3_

- [x] 4.3 Implement todo move endpoint
  - Create PATCH /api/todos/:id/move endpoint
  - Update todo status and position
  - Ensure user can only move their own todos
  - _Requirements: 4.3, 4.4, 9.4_

- [x] 4.4 Add error handling and validation
  - Implement request validation middleware
  - Add global error handler
  - Implement proper error responses
  - _Requirements: 6.4, 10.1, 10.2, 10.3_

- [x] 4.5 Write backend API tests
  - Test CRUD operations with test database
  - Test authorization (users can only access own todos)
  - Test validation and error handling
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 4.3, 4.4_

- [x] 5. Implement frontend authentication
- [x] 5.1 Create AuthContext and AuthProvider
  - Implement authentication state management
  - Create login, logout, and token refresh functions
  - Store tokens in memory (not localStorage)
  - _Requirements: 1.1, 1.4, 1.5, 7.1, 7.4_

- [x] 5.2 Create Login and LoginCallback pages
  - Implement login page with OAuth2 redirect
  - Implement callback page to handle authorization code
  - Handle OAuth2 errors and display messages
  - _Requirements: 1.1, 1.2, 10.1_

- [x] 5.3 Create ProtectedRoute component
  - Implement route protection with authentication check
  - Redirect unauthenticated users to login
  - Handle loading states
  - _Requirements: 1.1, 7.4_

- [x] 5.4 Setup Axios interceptors
  - Add Authorization header to all requests
  - Implement automatic token refresh on 401 errors
  - Handle network errors with retry logic
  - _Requirements: 1.5, 10.2, 10.4, 10.5_

- [x] 5.5 Write authentication component tests
  - Test AuthProvider state management
  - Test login and logout flows
  - Test token refresh logic
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 6. Implement frontend Todo UI components
- [x] 6.1 Create Header component
  - Display user profile information
  - Implement logout button
  - Add responsive design
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 7.1, 8.1, 8.2, 8.3_

- [x] 6.2 Create TodoForm component
  - Implement form for creating/editing todos
  - Add validation for title and description
  - Handle form submission and cancellation
  - _Requirements: 3.1, 3.4_

- [x] 6.3 Create TodoCard component
  - Display todo information (title, description)
  - Add edit and delete buttons
  - Implement responsive design
  - _Requirements: 3.4, 3.5, 8.1, 8.2, 8.3_

- [x] 6.4 Create TodoList component
  - Display list title and todo count
  - Render todo cards
  - Add "Add Todo" button for To Do list
  - _Requirements: 5.1, 5.2, 5.4, 5.5_

- [x] 6.5 Write UI component tests
  - Test component rendering
  - Test user interactions (buttons, forms)
  - Test responsive behavior
  - _Requirements: 2.1, 2.2, 2.3, 3.1, 3.4, 3.5_

- [x] 7. Implement drag-and-drop functionality
- [x] 7.1 Setup React DnD providers
  - Configure DnD context with HTML5 and Touch backends
  - Detect mobile devices for appropriate backend
  - Wrap TodoBoard with DndProvider
  - _Requirements: 4.1, 8.4, 8.5_

- [x] 7.2 Make TodoCard draggable
  - Implement useDrag hook in TodoCard
  - Add drag preview and visual feedback
  - Handle drag start and end events
  - _Requirements: 4.1, 4.2, 12.2_

- [x] 7.3 Make TodoList droppable
  - Implement useDrop hook in TodoList
  - Add drop zone highlighting
  - Handle drop events and update todo status
  - _Requirements: 4.2, 4.3, 4.4_

- [x] 7.4 Implement optimistic updates
  - Update UI immediately on drag-and-drop
  - Call API to persist changes
  - Rollback on error
  - _Requirements: 4.4, 6.1, 6.2, 6.3, 6.5, 12.3_

- [x] 7.5 Write drag-and-drop tests
  - Test drag and drop interactions
  - Test optimistic updates and rollback
  - Test touch gestures on mobile
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 8. Implement data fetching and state management
- [x] 8.1 Setup React Query
  - Configure QueryClient with caching settings
  - Wrap app with QueryClientProvider
  - Setup devtools for development
  - _Requirements: 12.4, 12.5_

- [x] 8.2 Create Todo API hooks
  - Implement useQuery hook for fetching todos
  - Implement useMutation hooks for create, update, delete
  - Add optimistic updates to mutations
  - Handle loading and error states
  - _Requirements: 3.2, 3.3, 6.1, 6.2, 6.3, 12.3_

- [x] 8.3 Implement TodoBoard component
  - Fetch todos with React Query
  - Group todos by status
  - Pass data to TodoList components
  - Handle loading and error states
  - _Requirements: 3.2, 5.3, 5.4_

- [x] 8.4 Write data fetching tests
  - Test React Query hooks with MSW
  - Test optimistic updates
  - Test error handling
  - _Requirements: 3.2, 3.3, 6.1, 6.2, 6.3_

- [x] 9. Implement security features
- [x] 9.1 Add CSRF protection
  - Implement CSRF token generation in backend
  - Add CSRF token to frontend requests
  - Validate CSRF tokens on state-changing operations
  - _Requirements: 11.3_

- [x] 9.2 Configure CORS properly
  - Set allowed origins from environment variables
  - Configure credentials and allowed headers
  - Test CORS in development and production
  - _Requirements: 11.4_

- [x] 9.3 Implement secure token storage
  - Store access tokens in memory only
  - Use HTTP-only cookies for refresh tokens
  - Clear tokens on logout
  - _Requirements: 11.2, 11.5, 7.5_

- [x] 9.4 Write security tests
  - Test CSRF protection
  - Test CORS configuration
  - Test token validation and authorization
  - _Requirements: 11.1, 11.2, 11.3_

- [x] 10. Implement responsive design
- [x] 10.1 Create responsive layout
  - Implement mobile layout (stacked lists)
  - Implement tablet layout (2 columns)
  - Implement desktop layout (3 columns)
  - _Requirements: 8.1, 8.2, 8.3, 8.4_

- [x] 10.2 Add touch support for mobile
  - Configure Touch backend for React DnD
  - Test drag-and-drop on mobile devices
  - Add touch-friendly button sizes
  - _Requirements: 8.5_

- [x] 10.3 Optimize for performance
  - Implement code splitting with React.lazy
  - Add loading skeletons
  - Optimize images and assets
  - _Requirements: 12.1, 12.2_

- [x] 10.4 Test responsive behavior
  - Test layouts at different breakpoints
  - Test touch interactions on mobile
  - Test performance metrics
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [-] 11. Add error handling and user feedback
- [x] 11.1 Create error boundary component
  - Implement React error boundary
  - Display user-friendly error messages
  - Add error reporting
  - _Requirements: 10.3_

- [x] 11.2 Add toast notifications
  - Implement toast notification system
  - Show success messages for operations
  - Show error messages with retry options
  - _Requirements: 6.4, 10.1, 10.2_

- [x] 11.3 Implement loading states
  - Add loading spinners for async operations
  - Implement skeleton screens for initial load
  - Show progress indicators for long operations
  - _Requirements: 12.1_

- [x] 11.4 Test error handling
  - Test error boundary
  - Test toast notifications
  - Test loading states
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [ ] 12. Setup development and production environments
- [ ] 12.1 Create Docker configuration
  - Create Dockerfile for frontend
  - Create Dockerfile for backend
  - Create docker-compose.yml for local development
  - _Requirements: All requirements (deployment)_

- [ ] 12.2 Configure environment variables
  - Create .env.example files
  - Document all environment variables
  - Setup different configs for dev/prod
  - _Requirements: All requirements (configuration)_

- [ ] 12.3 Add build scripts
  - Create production build scripts
  - Add linting and formatting scripts
  - Setup pre-commit hooks
  - _Requirements: All requirements (development)_

- [ ] 12.4 Create deployment documentation
  - Document deployment process
  - Create setup instructions
  - Add troubleshooting guide
  - _Requirements: All requirements (documentation)_

- [x] 13. End-to-end testing
  - Setup Cypress or Playwright
  - Write E2E tests for complete user flows
  - Test OAuth2 login flow
  - Test todo CRUD operations
  - Test drag-and-drop functionality
  - _Requirements: All requirements (testing)_

- [ ] 14. Performance optimization and monitoring
  - Add performance monitoring
  - Optimize bundle size
  - Implement lazy loading
  - Add analytics
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_
