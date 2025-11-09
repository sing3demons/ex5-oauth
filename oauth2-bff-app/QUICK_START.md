# Todo App Quick Start Guide

Get the Todo App with SSO up and running in minutes.

## Prerequisites

- Node.js 16+ and npm
- MongoDB running on `localhost:27017`
- OAuth2 Server running on `localhost:8080`
- jq (for JSON parsing)

## Step-by-Step Setup

### 1. Register OAuth2 Client

```bash
# From project root
./oauth2-bff-app/register_todo_client.sh
```

Copy the `client_id` and `client_secret` from the output.

### 2. Configure Backend

```bash
cd oauth2-bff-app/backend

# Copy environment template
cp .env.example .env

# Edit .env and add your client credentials
nano .env
```

Update these values in `.env`:
```bash
OAUTH2_CLIENT_ID=<your-client-id>
OAUTH2_CLIENT_SECRET=<your-client-secret>
```

### 3. Install Dependencies

```bash
# Backend
cd oauth2-bff-app/backend
npm install

# Frontend
cd ../frontend
npm install
```

### 4. Start the Application

```bash
# Terminal 1: Start backend
cd oauth2-bff-app/backend
npm run dev

# Terminal 2: Start frontend
cd oauth2-bff-app/frontend
npm run dev
```

### 5. Access the Application

Open your browser and navigate to:
```
http://localhost:3000
```

Click "Login" to authenticate via the OAuth2 server.

## Ports

- **Frontend:** http://localhost:3000
- **Backend:** http://localhost:4000
- **OAuth2 Server:** http://localhost:8080
- **MongoDB:** mongodb://localhost:27017

## Troubleshooting

### Client Registration Failed

Ensure the OAuth2 server is running:
```bash
./oauth2-server
```

### Backend Won't Start

Check that:
1. MongoDB is running
2. Port 4000 is available
3. Environment variables are set correctly

### Frontend Can't Connect

Verify:
1. Backend is running on port 4000
2. CORS is configured correctly
3. Frontend URL matches in backend `.env`

## Next Steps

- Read [CLIENT_REGISTRATION.md](./CLIENT_REGISTRATION.md) for detailed registration info
- Review [TODO_APP_README.md](./TODO_APP_README.md) for architecture details
- Check the [design document](../.kiro/specs/todo-app-with-sso/design.md) for implementation details

## Need Help?

See the full documentation:
- [Client Registration Guide](./CLIENT_REGISTRATION.md)
- [Setup Complete Guide](./SETUP_COMPLETE.md)
- [Todo App Guide](./TODO_APP_GUIDE.md)
