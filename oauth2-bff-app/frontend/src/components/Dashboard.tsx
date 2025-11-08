import { useAuth } from '../context/AuthContext';

export default function Dashboard() {
  const { user, logout, accessToken } = useAuth();

  if (!user) {
    return <div>Loading...</div>;
  }

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <div style={styles.header}>
          <h1 style={styles.title}>👋 Welcome!</h1>
          <button onClick={logout} style={styles.logoutButton}>
            Logout
          </button>
        </div>

        <div style={styles.userInfo}>
          <div style={styles.avatar}>
            {user.name?.charAt(0).toUpperCase() || user.email?.charAt(0).toUpperCase() || '?'}
          </div>
          <div>
            <h2 style={styles.name}>{user.name || 'User'}</h2>
            <p style={styles.email}>{user.email}</p>
          </div>
        </div>

        <div style={styles.section}>
          <h3 style={styles.sectionTitle}>🔐 Security Features</h3>
          <div style={styles.features}>
            <div style={styles.feature}>
              <span style={styles.icon}>✅</span>
              <div>
                <strong>Refresh Token</strong>
                <p style={styles.featureDesc}>Stored in HttpOnly cookie (not accessible to JavaScript)</p>
              </div>
            </div>
            <div style={styles.feature}>
              <span style={styles.icon}>✅</span>
              <div>
                <strong>Access Token</strong>
                <p style={styles.featureDesc}>Stored in memory only (cleared on page refresh)</p>
              </div>
            </div>
            <div style={styles.feature}>
              <span style={styles.icon}>✅</span>
              <div>
                <strong>Auto Refresh</strong>
                <p style={styles.featureDesc}>Token refreshes automatically before expiry</p>
              </div>
            </div>
            <div style={styles.feature}>
              <span style={styles.icon}>✅</span>
              <div>
                <strong>PKCE Protection</strong>
                <p style={styles.featureDesc}>Authorization code protected with PKCE</p>
              </div>
            </div>
          </div>
        </div>

        <div style={styles.section}>
          <h3 style={styles.sectionTitle}>📋 User Claims</h3>
          <pre style={styles.json}>
            {JSON.stringify(user, null, 2)}
          </pre>
        </div>

        <div style={styles.section}>
          <h3 style={styles.sectionTitle}>🔑 Access Token (First 50 chars)</h3>
          <code style={styles.token}>
            {accessToken?.substring(0, 50)}...
          </code>
          <p style={styles.note}>
            ⚠️ This token is stored in memory only and will be lost on page refresh.
            The refresh token in HttpOnly cookie will automatically get a new access token.
          </p>
        </div>

        <div style={styles.footer}>
          <p style={styles.footerText}>
            Try refreshing the page - you'll stay logged in! 🎉
          </p>
        </div>
      </div>
    </div>
  );
}

const styles = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    padding: '20px'
  },
  card: {
    maxWidth: '800px',
    margin: '0 auto',
    background: 'white',
    borderRadius: '16px',
    padding: '40px',
    boxShadow: '0 20px 60px rgba(0,0,0,0.3)'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '30px'
  },
  title: {
    fontSize: '32px',
    fontWeight: 'bold',
    margin: 0,
    color: '#333'
  },
  logoutButton: {
    padding: '10px 20px',
    fontSize: '14px',
    fontWeight: 'bold',
    color: 'white',
    background: '#dc3545',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer'
  },
  userInfo: {
    display: 'flex',
    alignItems: 'center',
    gap: '20px',
    padding: '20px',
    background: '#f8f9fa',
    borderRadius: '12px',
    marginBottom: '30px'
  },
  avatar: {
    width: '80px',
    height: '80px',
    borderRadius: '50%',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '32px',
    fontWeight: 'bold',
    color: 'white'
  },
  name: {
    fontSize: '24px',
    fontWeight: 'bold',
    margin: '0 0 5px 0',
    color: '#333'
  },
  email: {
    fontSize: '16px',
    color: '#666',
    margin: 0
  },
  section: {
    marginBottom: '30px'
  },
  sectionTitle: {
    fontSize: '20px',
    fontWeight: 'bold',
    marginBottom: '15px',
    color: '#333'
  },
  features: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '12px'
  },
  feature: {
    display: 'flex',
    gap: '12px',
    padding: '15px',
    background: '#f8f9fa',
    borderRadius: '8px'
  },
  icon: {
    fontSize: '20px'
  },
  featureDesc: {
    fontSize: '14px',
    color: '#666',
    margin: '5px 0 0 0'
  },
  json: {
    background: '#f8f9fa',
    padding: '15px',
    borderRadius: '8px',
    fontSize: '14px',
    overflow: 'auto',
    maxHeight: '200px'
  },
  token: {
    display: 'block',
    background: '#f8f9fa',
    padding: '15px',
    borderRadius: '8px',
    fontSize: '12px',
    fontFamily: 'monospace',
    wordBreak: 'break-all' as const,
    marginBottom: '10px'
  },
  note: {
    fontSize: '14px',
    color: '#666',
    margin: 0
  },
  footer: {
    textAlign: 'center' as const,
    paddingTop: '20px',
    borderTop: '1px solid #e0e0e0'
  },
  footerText: {
    fontSize: '16px',
    color: '#666',
    margin: 0
  }
};
