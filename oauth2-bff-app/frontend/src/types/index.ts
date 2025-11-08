export interface User {
  sub: string;
  email?: string;
  name?: string;
  email_verified?: boolean;
  [key: string]: any;
}

export interface AuthContextType {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: () => void;
  logout: () => void;
  refreshToken: () => Promise<void>;
}

export interface TokenInfo {
  access_token: string;
  expires_in: number;
  token_type: string;
}
