export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope?: string;
}

export interface UserInfo {
  sub: string;
  email?: string;
  name?: string;
  email_verified?: boolean;
  [key: string]: any;
}

export interface AuthSession {
  state: string;
  redirect_uri: string;
  nonce: string;
  timestamp: number;
}

export interface IDTokenClaims {
  iss: string;
  sub: string;
  aud: string | string[];
  exp: number;
  iat: number;
  nonce?: string;
  email?: string;
  name?: string;
  email_verified?: boolean;
  [key: string]: any;
}

declare global {
  namespace Express {
    interface Request {
      session?: AuthSession;
    }
  }
}
