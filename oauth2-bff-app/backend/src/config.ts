import dotenv from 'dotenv';
dotenv.config();
// const OAUTH2_SERVER = process.env.OAUTH2_SERVER_URL || 'http://localhost:8080';
// const CLIENT_ID = process.env.CLIENT_ID!;
// const CLIENT_SECRET = process.env.CLIENT_SECRET!;
// const FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:5173';
// const REDIRECT_URI = `${process.env.PORT ? `http://localhost:${process.env.PORT}` : 'http://localhost:3001'}/auth/callback`;
// const PORT = process.env.PORT || 3001;
// const FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:5173';
type Config = {
    PORT: number | string;
    FRONTEND_URL: string;
    OAUTH2_SERVER: string;
    CLIENT_ID: string;
    CLIENT_SECRET: string;
    REDIRECT_URI: string;
};

const config: Config = {
    PORT: process.env.PORT || 3001,
    OAUTH2_SERVER: process.env.OAUTH2_SERVER_URL || 'http://localhost:8080',
    CLIENT_ID: process.env.CLIENT_ID!,
    CLIENT_SECRET: process.env.CLIENT_SECRET!,
    FRONTEND_URL: process.env.FRONTEND_URL || 'http://localhost:5173',
    REDIRECT_URI: `${process.env.PORT ? `http://localhost:${process.env.PORT}` : 'http://localhost:3001'}/auth/callback`,
};


export default config;