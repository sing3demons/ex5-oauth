import dotenv from 'dotenv';
dotenv.config();
type Config = {
    PORT: number | string;
    FRONTEND_URL: string;
    OAUTH2_SERVER: string;
    CLIENT_ID: string;
    CLIENT_SECRET: string;
    REDIRECT_URI: string;
    CORS_ORIGINS: string[];
};


const config: Config = {
    PORT: process.env.PORT || 3001,
    OAUTH2_SERVER: process.env.OAUTH2_SERVER_URL || 'http://localhost:8080',
    CLIENT_ID: process.env.OAUTH2_CLIENT_ID!,
    CLIENT_SECRET: process.env.OAUTH2_CLIENT_SECRET!,
    FRONTEND_URL: process.env.FRONTEND_URL || 'http://localhost:3000',
    REDIRECT_URI: process.env.OAUTH2_REDIRECT_URI || `http://localhost:${process.env.PORT || 4000}/auth/callback`,
    CORS_ORIGINS: process.env.CORS_ORIGIN 
        ? process.env.CORS_ORIGIN.split(',').map(origin => origin.trim())
        : [
            process.env.FRONTEND_URL || 'http://localhost:3000',
            process.env.OAUTH2_SERVER_URL || 'http://localhost:8080'
          ],
};


export default config;