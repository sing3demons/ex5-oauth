import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function LoginCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleCallback = async () => {
      // Check for OAuth2 errors
      const errorParam = searchParams.get('error');
      const errorDescription = searchParams.get('error_description');

      if (errorParam) {
        const errorMessage = getErrorMessage(errorParam, errorDescription);
        setError(errorMessage);
        return;
      }

      // Check for access token (from backend redirect)
      const accessToken = searchParams.get('access_token');
      const expiresIn = searchParams.get('expires_in');

      if (accessToken && expiresIn) {
        // Token will be handled by AuthContext
        // Wait a moment for AuthContext to process
        setTimeout(() => {
          if (isAuthenticated) {
            navigate('/dashboard', { replace: true });
          }
        }, 100);
      } else if (!errorParam) {
        // No token and no error - something went wrong
        setError('Authentication failed. No token received.');
      }
    };

    handleCallback();
  }, [searchParams, navigate, isAuthenticated]);

  const getErrorMessage = (error: string, description: string | null): string => {
    const errorMessages: Record<string, string> = {
      access_denied: 'You denied access to the application.',
      invalid_request: 'Invalid authentication request.',
      invalid_state: 'Invalid state parameter. Please try again.',
      session_expired: 'Your session has expired. Please try again.',
      token_exchange_failed: 'Failed to exchange authorization code for tokens.',
      invalid_id_token: 'Invalid ID token received.',
    };

    const message = errorMessages[error] || 'An unknown error occurred during authentication.';
    return description ? `${message} (${description})` : message;
  };

  const handleRetry = () => {
    setError(null);
    navigate('/login', { replace: true });
  };

  if (error) {
    return (
      <div style={styles.container}>
        <div style={styles.card}>
          <div style={styles.errorIcon}>⚠️</div>
          <h1 style={styles.title}>Authentication Error</h1>
          <p style={styles.errorMessage}>{error}</p>
          <button onClick={handleRetry} style={styles.button}>
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <div style={styles.spinner}></div>
        <h2 style={styles.title}>Completing authentication...</h2>
        <p style={styles.subtitle}>Please wait while we log you in.</p>
      </div>
    </div>
  );
}

const styles = {
  container: {
    minHeight: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    padding: '20px',
  },
  card: {
    background: 'white',
    borderRadius: '16px',
    padding: '40px',
    maxWidth: '500px',
    width: '100%',
    boxShadow: '0 20px 60px rgba(0,0,0,0.3)',
    textAlign: 'center' as const,
  },
  spinner: {
    width: '50px',
    height: '50px',
    border: '5px solid rgba(102, 126, 234, 0.3)',
    borderTop: '5px solid #667eea',
    borderRadius: '50%',
    animation: 'spin 1s linear infinite',
    margin: '0 auto 20px',
  },
  errorIcon: {
    fontSize: '64px',
    marginBottom: '20px',
  },
  title: {
    fontSize: '24px',
    fontWeight: 'bold',
    marginBottom: '16px',
    color: '#333',
  },
  subtitle: {
    fontSize: '16px',
    color: '#666',
    margin: 0,
  },
  errorMessage: {
    fontSize: '16px',
    color: '#ef4444',
    marginBottom: '24px',
    lineHeight: '1.5',
  },
  button: {
    padding: '12px 32px',
    fontSize: '16px',
    fontWeight: 'bold',
    color: 'white',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'transform 0.2s',
  },
};
