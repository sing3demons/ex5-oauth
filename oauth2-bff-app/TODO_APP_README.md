# Todo App with OAuth2/OIDC SSO

A modern, responsive Todo application with OAuth2/OIDC Single Sign-On integration and drag-and-drop functionality.

## Project Structure

```
oauth2-bff-app/
├── backend/          # Express backend with TypeScript
├── frontend/         # React frontend with TypeScript
└── README.md         # This file
```

## Technology Stack

### Frontend
- React 18+ with TypeScript
- React DnD (Drag and Drop)
- React Query for data fetching
- React Router for navigation
- Tailwind CSS for styling
- Axios for HTTP requests
- Vite for build tooling

### Backend
- Node.js with Express
- TypeScript
- MongoDB with Mongoose
- Express Session for session management
- Helmet for security headers
- CORS for cross-origin requests
- JWT for token validation

## Prerequisites

- Node.js 18+ and npm
- MongoDB 5.0+
- OAuth2 Server running on port 8080

## Setup Instructions

### 1. Install Dependencies

```bash
# Install backend dependencies
cd backend
npm install

# Install frontend dependencies
cd ../frontend
npm install
```

### 2. Configure Environment Variables

```bash
# Backend configuration
cd backend
cp .env.example .env
# Edit .env with your OAuth2 client credentials and MongoDB URI

# Frontend configuration
cd ../frontend
cp .env.example .env
# Edit .env if needed (defaults should work for local development)
```

### 3. Register OAuth2 Client

Before running the application, register a client with the OAuth2 server:

```bash
# From the oauth2-bff-app directory
./register_bff_client.sh
```

Update the `OAUTH2_CLIENT_ID` and `OAUTH2_CLIENT_SECRET` in `backend/.env` with the credentials returned.

### 4. Start MongoDB

```bash
# Make sure MongoDB is running
mongod --dbpath /path/to/data
```

### 5. Start the Applications

```bash
# Terminal 1: Start backend
cd backend
npm run dev

# Terminal 2: Start frontend
cd frontend
npm run dev
```

The frontend will be available at http://localhost:3000
The backend API will be available at http://localhost:4000

## Development

### Backend Development

```bash
cd backend

# Run in development mode with hot reload
npm run dev

# Type checking
npm run type-check

# Build for production
npm run build

# Start production server
npm start

# Linting
npm run lint

# Format code
npm run format
```

### Frontend Development

```bash
cd frontend

# Run in development mode with hot reload
npm run dev

# Type checking
npm run type-check

# Build for production
npm run build

# Preview production build
npm run preview

# Linting
npm run lint

# Format code
npm run format
```

## Features

- ✅ OAuth2/OIDC Single Sign-On authentication
- ✅ Drag-and-drop todo management
- ✅ Three-column board (To Do, In Progress, Done)
- ✅ Responsive design (mobile, tablet, desktop)
- ✅ Real-time updates with optimistic UI
- ✅ Secure token management
- ✅ Session management with automatic token refresh
- ✅ User profile display

## Architecture

The application follows a Backend-for-Frontend (BFF) pattern:

1. **Frontend** - React SPA that communicates with the backend API
2. **Backend** - Express API that handles OAuth2 flow and proxies requests
3. **OAuth2 Server** - Existing OAuth2/OIDC server for authentication
4. **MongoDB** - Database for storing todo items

## Security Features

- HTTP-only cookies for refresh tokens
- CSRF protection
- Secure session management
- Token validation with OAuth2 server
- CORS configuration
- Helmet security headers

## API Documentation

See the design document at `.kiro/specs/todo-app-with-sso/design.md` for detailed API documentation.

## Troubleshooting

### MongoDB Connection Issues
- Ensure MongoDB is running: `mongod --version`
- Check the connection string in `backend/.env`

### OAuth2 Authentication Issues
- Verify the OAuth2 server is running on port 8080
- Check client credentials in `backend/.env`
- Ensure redirect URI matches: `http://localhost:4000/auth/callback`

### Port Conflicts
- Backend default: 4000 (change in `backend/.env`)
- Frontend default: 3000 (change in `frontend/vite.config.ts`)

## License

MIT
