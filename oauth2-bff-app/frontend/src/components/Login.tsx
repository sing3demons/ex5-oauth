import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Login() {
  const { login, isAuthenticated } = useAuth();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Redirect if already authenticated
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }

    // Check for error in URL
    const errorParam = searchParams.get('error');
    if (errorParam) {
      const errorMessages: Record<string, string> = {
        access_denied: 'Access was denied. Please try again.',
        invalid_request: 'Invalid request. Please try again.',
        server_error: 'Server error occurred. Please try again later.',
      };
      setError(errorMessages[errorParam] || 'An error occurred during login.');
    }
  }, [isAuthenticated, navigate, searchParams]);

  const handleLogin = () => {
    setError(null);
    login();
  };

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h1 style={styles.title}>🔐 Todo App with SSO</h1>
        <p style={styles.subtitle}>
          Secure authentication with OAuth2/OIDC
        </p>
        
        {error && (
          <div style={styles.errorBanner}>
            <span style={styles.errorIcon}>⚠️</span>
            <span>{error}</span>
          </div>
        )}
        
        <div style={styles.features}>
          <div style={styles.feature}>
            <span style={styles.icon}>✅</span>
            <span>HttpOnly Cookies</span>
          </div>
          <div style={styles.feature}>
            <span style={styles.icon}>✅</span>
            <span>PKCE Flow</span>
          </div>
          <div style={styles.feature}>
            <span style={styles.icon}>✅</span>
            <span>Auto Token Refresh</span>
          </div>
          <div style={styles.feature}>
            <span style={styles.icon}>✅</span>
            <span>Memory-only Access Tokens</span>
          </div>
        </div>

        <button onClick={handleLogin} style={styles.button}>
          Login with OAuth2
        </button>

        <p style={styles.note}>
          You'll be redirected to the OAuth2 server for authentication
        </p>
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
    padding: '20px'
  },
  card: {
    background: 'white',
    borderRadius: '16px',
    padding: '40px',
    maxWidth: '500px',
    width: '100%',
    boxShadow: '0 20px 60px rgba(0,0,0,0.3)'
  },
  title: {
    fontSize: '32px',
    fontWeight: 'bold',
    marginBottom: '10px',
    textAlign: 'center' as const,
    color: '#333'
  },
  subtitle: {
    fontSize: '16px',
    color: '#666',
    textAlign: 'center' as const,
    marginBottom: '30px'
  },
  features: {
    marginBottom: '30px'
  },
  feature: {
    display: 'flex',
    alignItems: 'center',
    padding: '12px',
    marginBottom: '8px',
    background: '#f8f9fa',
    borderRadius: '8px',
    fontSize: '14px',
    color: '#333'
  },
  icon: {
    marginRight: '12px',
    fontSize: '18px'
  },
  button: {
    width: '100%',
    padding: '16px',
    fontSize: '16px',
    fontWeight: 'bold',
    color: 'white',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'transform 0.2s',
    marginBottom: '16px'
  },
  note: {
    fontSize: '12px',
    color: '#999',
    textAlign: 'center' as const,
    margin: 0
  },
  errorBanner: {
    display: 'flex',
    alignItems: 'center',
    padding: '12px 16px',
    marginBottom: '20px',
    background: '#fee',
    border: '1px solid #fcc',
    borderRadius: '8px',
    color: '#c33',
    fontSize: '14px',
  },
  errorIcon: {
    marginRight: '8px',
    fontSize: '18px',
  }
};
