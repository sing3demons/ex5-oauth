# 📝 Todo App with OAuth2 BFF - Quick Start Guide

## 🎯 Features

✅ **Full CRUD Operations** - Create, Read, Update, Delete todos  
✅ **Drag & Drop** - Move tasks between columns (Todo → In Progress → Done)  
✅ **Beautiful UI** - Modern gradient design with smooth animations  
✅ **Priority Levels** - Low, Medium, High with color coding  
✅ **User-specific** - Each user sees only their own todos  
✅ **MongoDB Storage** - Persistent data storage  
✅ **Secure** - Protected by OAuth2 + OIDC authentication  

## 🚀 Quick Start

### 1. Start MongoDB
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

### 2. Start OAuth2 Server (Go)
```bash
cd ..
go run main.go
```

### 3. Start BFF Backend
```bash
cd oauth2-bff-app/backend
npm install
npm run dev
```

### 4. Start React Frontend
```bash
cd oauth2-bff-app/frontend
npm install
npm run dev
```

### 5. Open Browser
```
http://localhost:5173
```

## 📱 How to Use

### Login
1. Click "Login with OAuth2"
2. Enter credentials on OAuth2 server
3. You'll be redirected back to the Todo app

### Create Todo
1. Click "+ New Task" button
2. Enter title (required)
3. Add description (optional)
4. Select priority (Low/Medium/High)
5. Click "Create Task"

### Drag & Drop
1. Click and hold a todo card
2. Drag it to another column
3. Release to drop
4. Status updates automatically!

### Edit Todo
1. Click the ✏️ (edit) icon on a card
2. Modify title or description
3. Click "Save"

### Delete Todo
1. Click the 🗑️ (delete) icon on a card
2. Confirm deletion

## 🎨 UI Features

### Columns
- **📋 To Do** - New tasks start here
- **🚀 In Progress** - Tasks you're working on
- **✅ Done** - Completed tasks

### Priority Colors
- 🟢 **Low** - Green
- 🟡 **Medium** - Orange  
- 🔴 **High** - Red

### Stats Dashboard
- Real-time count of tasks in each column
- Beautiful gradient cards

## 🔐 Security

- ✅ All API endpoints protected by OAuth2
- ✅ User-specific data isolation
- ✅ HttpOnly cookies for refresh tokens
- ✅ OIDC compliant authentication
- ✅ Auto token refresh

## 🗄️ Database Schema

### MongoDB Collection: `todos`

```typescript
{
  id: string;           // UUID
  userId: string;       // From OAuth2 token
  title: string;        // Task title
  description?: string; // Optional details
  status: 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high';
  createdAt: Date;
  updatedAt: Date;
}
```

### Indexes
- `userId + createdAt` (descending)

## 📡 API Endpoints

### Get All Todos
```bash
GET /api/todos
Authorization: Bearer {access_token}
```

### Create Todo
```bash
POST /api/todos
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "title": "My Task",
  "description": "Task details",
  "priority": "high"
}
```

### Update Todo
```bash
PUT /api/todos/:id
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "title": "Updated Title",
  "status": "in_progress"
}
```

### Update Status (Drag & Drop)
```bash
PATCH /api/todos/:id/status
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "status": "done"
}
```

### Delete Todo
```bash
DELETE /api/todos/:id
Authorization: Bearer {access_token}
```

## 🎭 Demo Workflow

1. **Login** → Authenticate with OAuth2
2. **Create** → Add "Build Todo App" (High priority)
3. **Drag** → Move to "In Progress"
4. **Create** → Add "Test Drag & Drop" (Medium priority)
5. **Edit** → Update description
6. **Drag** → Move first task to "Done"
7. **Delete** → Remove completed task
8. **Logout** → Clear session

## 🐛 Troubleshooting

### MongoDB Connection Error
```bash
# Check if MongoDB is running
docker ps | grep mongodb

# Restart MongoDB
docker restart mongodb
```

### Port Already in Use
```bash
# Kill process on port 3001 (BFF)
lsof -ti:3001 | xargs kill -9

# Kill process on port 5173 (Frontend)
lsof -ti:5173 | xargs kill -9
```

### "Failed to fetch todos"
- Check if BFF server is running
- Check if access token is valid
- Check MongoDB connection

### Drag & Drop Not Working
- Make sure @dnd-kit packages are installed
- Check browser console for errors
- Try refreshing the page

## 🎨 Customization

### Change Colors
Edit `frontend/src/components/TodoBoard.tsx`:
```typescript
background: 'linear-gradient(135deg, #YOUR_COLOR_1 0%, #YOUR_COLOR_2 100%)'
```

### Add More Columns
1. Add new status to `types/todo.ts`
2. Add column in `TodoBoard.tsx`
3. Update backend validation

### Change Priority Levels
1. Update `types/todo.ts`
2. Update `CreateTodoModal.tsx`
3. Update `TodoCard.tsx` colors

## 📚 Tech Stack

### Backend
- Node.js + TypeScript
- Express.js
- MongoDB
- OAuth2 + OIDC

### Frontend
- React 18 + TypeScript
- @dnd-kit (Drag & Drop)
- Axios
- Vite

## 🚀 Production Deployment

### Environment Variables

**Backend:**
```env
NODE_ENV=production
MONGODB_URI=mongodb://your-mongo-host:27017
MONGODB_DB=oauth2_bff_app_prod
OAUTH2_SERVER_URL=https://oauth.yourdomain.com
CLIENT_ID=your_prod_client_id
CLIENT_SECRET=your_prod_client_secret
FRONTEND_URL=https://app.yourdomain.com
```

**Frontend:**
```env
VITE_BFF_URL=https://bff.yourdomain.com
```

### Build
```bash
# Backend
cd backend
npm run build
npm start

# Frontend
cd frontend
npm run build
# Serve dist/ folder with nginx or similar
```

## 🎉 Enjoy!

You now have a fully functional Todo app with:
- ✅ Beautiful drag & drop interface
- ✅ Secure OAuth2 authentication
- ✅ Persistent MongoDB storage
- ✅ User-specific data
- ✅ Auto token refresh

Happy task managing! 🚀
