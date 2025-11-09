import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import axios from 'axios';
import { User, AuthContextType, TokenInfo } from '../types';
import { setAccessToken as setApiAccessToken, fetchCsrfToken, setCsrfToken } from '../services/api';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const refreshTimerRef = useRef<number | null>(null);

  // Update API token whenever accessToken changes
  useEffect(() => {
    setApiAccessToken(accessToken);
  }, [accessToken]);

  // Auto-refresh token before expiry
  const scheduleTokenRefresh = useCallback((expiresIn: number) => {
    // Clear existing timer
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current);
    }

    // Refresh 1 minute before expiry
    const refreshTime = (expiresIn - 60) * 1000;
    
    if (refreshTime > 0) {
      refreshTimerRef.current = setTimeout(async () => {
        try {
          await refreshToken();
        } catch (error) {
          console.error('Auto-refresh failed:', error);
          logout();
        }
      }, refreshTime);
    }
  }, []);

  // Refresh access token
  const refreshToken = useCallback(async () => {
    try {
      const response = await axios.post<TokenInfo>(
        `${API_URL}/auth/refresh`,
        {},
        { withCredentials: true }
      );

      const { access_token, expires_in } = response.data;
      setAccessToken(access_token);
      
      // Schedule next refresh
      scheduleTokenRefresh(expires_in);
      
      // Fetch user info with new token
      await fetchUserInfo(access_token);
    } catch (error) {
      console.error('Token refresh failed:', error);
      throw error;
    }
  }, [scheduleTokenRefresh]);

  // Fetch user info
  const fetchUserInfo = async (token: string) => {
    try {
      const response = await axios.get<User>(
        `${API_URL}/auth/userinfo`,
        {
          headers: { Authorization: `Bearer ${token}` },
          withCredentials: true
        }
      );
      setUser(response.data);
    } catch (error) {
      console.error('Failed to fetch user info:', error);
      throw error;
    }
  };

  // Login - redirect to backend
  const login = useCallback(async () => {
    try {
      const response = await axios.get<{ authorization_url: string }>(
        `${API_URL}/auth/login`,
        { withCredentials: true }
      );
      
      // Redirect to OAuth2 authorization
      window.location.href = response.data.authorization_url;
    } catch (error) {
      console.error('Login failed:', error);
    }
  }, []);

  // Logout
  const logout = useCallback(async () => {
    try {
      // Clear refresh timer
      if (refreshTimerRef.current) {
        clearTimeout(refreshTimerRef.current);
      }

      // Call backend logout
      await axios.post(
        `${API_URL}/auth/logout`,
        {},
        { withCredentials: true }
      );

      // Clear state
      setAccessToken(null);
      setUser(null);
      
      // Clear CSRF token
      setCsrfToken(null);
      
      // Broadcast logout to other tabs
      localStorage.setItem('logout', Date.now().toString());
    } catch (error) {
      console.error('Logout failed:', error);
    }
  }, []);

  // Handle OAuth callback and initial token check
  useEffect(() => {
    const initializeAuth = async () => {
      // Fetch CSRF token on initialization
      try {
        await fetchCsrfToken();
      } catch (error) {
        console.error('Failed to fetch CSRF token on init:', error);
      }

      const params = new URLSearchParams(window.location.search);
      const accessTokenParam = params.get('access_token');
      const expiresInParam = params.get('expires_in');

      if (accessTokenParam && expiresInParam) {
        // Store token from callback
        setAccessToken(accessTokenParam);
        
        // Schedule refresh
        scheduleTokenRefresh(parseInt(expiresInParam));
        
        // Fetch user info
        fetchUserInfo(accessTokenParam).finally(() => {
          setIsLoading(false);
        });

        // Clean URL
        window.history.replaceState({}, document.title, window.location.pathname);
      } else {
        // Try to refresh token on mount (check for existing session)
        refreshToken()
          .catch(() => {
            // No valid refresh token
            setIsLoading(false);
          });
      }
    };

    initializeAuth();
  }, [scheduleTokenRefresh, refreshToken]);

  // Listen for logout events from other tabs
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'logout') {
        setAccessToken(null);
        setUser(null);
        setCsrfToken(null);
        if (refreshTimerRef.current) {
          clearTimeout(refreshTimerRef.current);
        }
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (refreshTimerRef.current) {
        clearTimeout(refreshTimerRef.current);
      }
    };
  }, []);

  const value: AuthContextType = {
    user,
    accessToken,
    isAuthenticated: !!accessToken && !!user,
    isLoading,
    login,
    logout,
    refreshToken
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
