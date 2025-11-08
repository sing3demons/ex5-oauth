import { useState, useEffect } from 'react'
import { useSSO } from '../context/SSOContext'

function AppA() {
  const { login, logout, isLoggedIn, users, tokens } = useSSO()
  const [email, setEmail] = useState('test@example.com')
  const [password, setPassword] = useState('password123')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [success, setSuccess] = useState(null)

  const appId = 'app-a'
  const loggedIn = isLoggedIn(appId)
  const user = users[appId]
  const token = tokens[appId]

  const handleLogin = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    setSuccess(null)

    try {
      await login(appId, email, password)
      setSuccess('✅ Login successful! You can now access App B and App C without logging in again.')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    logout()
    setSuccess('Logged out from all apps')
  }

  if (loading) {
    return <div className="loading">Logging in...</div>
  }

  if (loggedIn) {
    return (
      <div>
        <div className="app-header">
          <h2>📱 App A - E-commerce</h2>
          <span className="status-badge logged-in">Logged In</span>
        </div>

        {success && <div className="success">{success}</div>}

        <div className="user-info">
          <h3>Welcome, {user?.name}!</h3>
          <p><strong>Email:</strong> {user?.email}</p>
          <p><strong>User ID:</strong> {user?.sub}</p>
        </div>

        <div className="sso-demo">
          <h3>🎉 SSO is Active!</h3>
          <p>You're now logged into App A. Try switching to App B or App C - they will automatically get tokens using Token Exchange without requiring you to login again!</p>
        </div>

        <div className="button-group">
          <button className="btn btn-danger" onClick={handleLogout}>
            Logout from All Apps
          </button>
        </div>

        <details style={{ marginTop: '20px' }}>
          <summary style={{ cursor: 'pointer', fontWeight: 'bold' }}>
            🔍 View Access Token
          </summary>
          <div className="token-display">
            {token?.access_token}
          </div>
        </details>
      </div>
    )
  }

  return (
    <div>
      <div className="app-header">
        <h2>📱 App A - E-commerce</h2>
        <span className="status-badge logged-out">Not Logged In</span>
      </div>

      {error && <div className="error">{error}</div>}
      {success && <div className="success">{success}</div>}

      <div className="sso-demo">
        <h3>👋 Welcome to App A!</h3>
        <p>This is the first app. Login here to enable SSO for App B and App C.</p>
      </div>

      <form className="login-form" onSubmit={handleLogin}>
        <div className="form-group">
          <label>Email:</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <label>Password:</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>

        <button type="submit" className="btn btn-primary" style={{ width: '100%' }}>
          Login to App A
        </button>
      </form>

      <div style={{ marginTop: '20px', padding: '15px', background: '#f5f5f5', borderRadius: '4px' }}>
        <p style={{ fontSize: '14px', color: '#666' }}>
          <strong>Test Credentials:</strong><br/>
          Email: test@example.com<br/>
          Password: password123
        </p>
        <p style={{ fontSize: '12px', color: '#999', marginTop: '10px' }}>
          Note: You need to register this user first using the OAuth server's register endpoint.
        </p>
      </div>
    </div>
  )
}

export default AppA
