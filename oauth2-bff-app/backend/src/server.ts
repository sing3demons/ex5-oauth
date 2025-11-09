import express from 'express';
import cors from 'cors';
import cookieParser from 'cookie-parser';
import helmet from 'helmet';
import dotenv from 'dotenv';
import authRoutes from './routes/auth';
import { connectDB } from './db/mongodb';
import config from './config';
import { csrfProtection } from './middleware/csrf';

// Load environment variables
dotenv.config();

const app = express();

// Security middleware
app.use(helmet());

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Cookie parser with secret for signed cookies
app.use(cookieParser(process.env.SESSION_SECRET || 'change-this-secret-in-production'));

// CORS configuration - Properly configured with environment variables
console.log('🔒 CORS allowed origins:', config.CORS_ORIGINS);

app.use(cors({
  origin: (origin, callback) => {
    // Allow requests with no origin (like mobile apps or curl requests)
    if (!origin) {
      console.log('✅ CORS: Allowing request with no origin');
      return callback(null, true);
    }
    
    // For development: allow localhost origins
    if (origin.startsWith('http://localhost:') || origin.startsWith('http://127.0.0.1:')) {
      console.log(`✅ CORS: Allowing localhost origin: ${origin}`);
      return callback(null, true);
    }
    
    // Check if origin is in allowed list
    if (config.CORS_ORIGINS.includes(origin)) {
      console.log(`✅ CORS: Allowing origin: ${origin}`);
      callback(null, true);
    } else {
      console.warn(`❌ CORS blocked origin: ${origin}`);
      console.warn(`   Allowed origins: ${config.CORS_ORIGINS.join(', ')}`);
      callback(new Error('Not allowed by CORS'));
    }
  },
  credentials: true, // Allow cookies and authorization headers
  methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-CSRF-Token'],
  exposedHeaders: ['X-CSRF-Token'],
  maxAge: 86400, // Cache preflight requests for 24 hours
}));

// Endpoint to get CSRF token
app.get('/csrf-token', csrfProtection, (req, res) => {
  res.json({ csrfToken: req.csrfToken() });
});

// Routes
app.use('/auth', authRoutes);

// Todo routes
import todoRoutes from './routes/todos';
app.use('/api/todos', todoRoutes);

// Health check
app.get('/health', (_req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString()
  });
});

// Error handling
import { errorHandler, notFoundHandler } from './middleware/errorHandler';

// 404 handler (must be after all routes)
app.use(notFoundHandler);

// Global error handler (must be last)
app.use(errorHandler);

// Start server
async function startServer() {
  try {
    // Connect to MongoDB
    await connectDB();
    
    app.listen(config.PORT, () => {
      console.log(`🚀 BFF Server running on http://localhost:${config.PORT}`);
      console.log(`📱 Frontend URL: ${config.FRONTEND_URL}`);
      console.log(`🔐 OAuth2 Server: ${config.OAUTH2_SERVER}`);
      console.log(`🆔 Client ID: ${config.CLIENT_ID}`);
    });
  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

startServer();
