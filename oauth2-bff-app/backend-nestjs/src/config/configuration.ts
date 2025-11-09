export interface OAuth2Config {
  serverUrl: string;
  clientId: string;
  clientSecret: string;
}

export interface FrontendConfig {
  url: string;
}

export interface DatabaseConfig {
  uri: string;
  name: string;
}

export interface AppConfig {
  port: number;
  nodeEnv: string;
  oauth2: OAuth2Config;
  frontend: FrontendConfig;
  database: DatabaseConfig;
}

export default (): AppConfig => ({
  port: parseInt(process.env.PORT || '3001', 10),
  nodeEnv: process.env.NODE_ENV || 'development',
  oauth2: {
    serverUrl: process.env.OAUTH2_SERVER_URL || '',
    clientId: process.env.CLIENT_ID || '',
    clientSecret: process.env.CLIENT_SECRET || '',
  },
  frontend: {
    url: process.env.FRONTEND_URL || '',
  },
  database: {
    uri: process.env.MONGODB_URI || '',
    name: process.env.MONGODB_DB || '',
  },
});
