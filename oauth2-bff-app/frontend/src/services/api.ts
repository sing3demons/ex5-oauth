import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000';

// Create axios instance
const api = axios.create({
  baseURL: API_URL,
  withCredentials: true, // Important for cookies
});

/**
 * Secure Token Storage
 * 
 * Security measures:
 * 1. Access tokens are stored in memory only (not localStorage/sessionStorage)
 *    - Prevents XSS attacks from stealing tokens
 *    - Tokens are lost on page refresh (requires refresh token)
 * 
 * 2. Refresh tokens are stored in HTTP-only cookies by the backend
 *    - Cannot be accessed by JavaScript (prevents XSS)
 *    - Automatically sent with requests to backend
 *    - Secure flag enabled in production (HTTPS only)
 * 
 * 3. CSRF tokens are stored in memory and sent as headers
 *    - Protects against CSRF attacks on state-changing operations
 *    - Automatically refreshed when invalid
 */
let accessToken: string | null = null;
let csrfToken: string | null = null;

export const setAccessToken = (token: string | null) => {
  accessToken = token;
};

export const getAccessToken = () => accessToken;

export const setCsrfToken = (token: string | null) => {
  csrfToken = token;
};

export const getCsrfToken = () => csrfToken;

// Fetch CSRF token from server
export const fetchCsrfToken = async (): Promise<string> => {
  try {
    const response = await axios.get(`${API_URL}/csrf-token`, {
      withCredentials: true,
    });
    const token = response.data.csrfToken;
    setCsrfToken(token);
    return token;
  } catch (error) {
    console.error('Failed to fetch CSRF token:', error);
    throw error;
  }
};

// Request interceptor - Add Authorization and CSRF headers
api.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    if (accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    
    // Add CSRF token for state-changing operations
    const method = config.method?.toLowerCase();
    if (method && ['post', 'put', 'patch', 'delete'].includes(method)) {
      // Fetch CSRF token if not already available
      if (!csrfToken) {
        try {
          await fetchCsrfToken();
        } catch (error) {
          console.error('Failed to fetch CSRF token:', error);
        }
      }
      
      if (csrfToken && config.headers) {
        config.headers['X-CSRF-Token'] = csrfToken;
      }
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - Handle token refresh and errors
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: Error | null, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });

  failedQueue = [];
};

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
      _retryCount?: number;
    };

    // Handle 403 errors (CSRF token invalid/missing)
    if (error.response?.status === 403 && originalRequest && !originalRequest._retry) {
      const errorData = error.response.data as { code?: string };
      
      // If CSRF error, fetch new token and retry
      if (errorData?.code === 'EBADCSRFTOKEN') {
        originalRequest._retry = true;
        
        try {
          await fetchCsrfToken();
          
          // Retry the original request with new CSRF token
          if (originalRequest.headers && csrfToken) {
            originalRequest.headers['X-CSRF-Token'] = csrfToken;
          }
          
          return api(originalRequest);
        } catch (csrfError) {
          return Promise.reject(csrfError);
        }
      }
    }

    // Handle 401 errors (token expired)
    if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
      if (isRefreshing) {
        // If already refreshing, queue this request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return api(originalRequest);
          })
          .catch((err) => {
            return Promise.reject(err);
          });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Attempt to refresh token
        const response = await axios.post(
          `${API_URL}/auth/refresh`,
          {},
          { withCredentials: true }
        );

        const newToken = response.data.access_token;
        setAccessToken(newToken);

        // Update the failed request with new token
        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
        }

        processQueue(null, newToken);
        isRefreshing = false;

        // Retry the original request
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError as Error, null);
        isRefreshing = false;

        // Clear token and redirect to login
        setAccessToken(null);
        window.location.href = '/login';

        return Promise.reject(refreshError);
      }
    }

    // Handle network errors with retry logic
    if (!error.response && originalRequest) {
      originalRequest._retryCount = originalRequest._retryCount || 0;

      // Retry up to 2 times for network errors
      if (originalRequest._retryCount < 2) {
        originalRequest._retryCount += 1;

        // Wait before retrying (exponential backoff)
        const delay = Math.pow(2, originalRequest._retryCount) * 1000;
        await new Promise((resolve) => setTimeout(resolve, delay));

        return api(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);

export default api;
