#!/bin/bash

echo "🚀 Setting up OAuth2 BFF Application..."
echo ""

# Backend setup
echo "📦 Setting up Backend..."
cd backend
cp .env.example .env
npm install
echo "✅ Backend setup complete"
echo ""

# Frontend setup
echo "📦 Setting up Frontend..."
cd ../frontend
cp .env.example .env
npm install
echo "✅ Frontend setup complete"
echo ""

echo "✨ Setup complete!"
echo ""
echo "📝 Next steps:"
echo "1. Make sure OAuth2 server is running on http://localhost:8080"
echo "2. Start BFF server: cd backend && npm run dev"
echo "3. Start Frontend: cd frontend && npm run dev"
echo "4. Open http://localhost:5173 in your browser"
echo ""
echo "🔐 Security Features:"
echo "  ✅ HttpOnly Cookies for refresh tokens"
echo "  ✅ PKCE flow for authorization"
echo "  ✅ Auto token refresh"
echo "  ✅ Memory-only access tokens"
echo "  ✅ Multi-tab logout sync"
