import { useState, useEffect } from 'react'
import { useSSO } from '../context/SSOContext'

function AppB() {
  const { getTokenForApp, logout, isLoggedIn, users, tokens, hasAnyToken } = useSSO()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const appId = 'app-b'
  const loggedIn = isLoggedIn(appId)
  const user = users[appId]
  const token = tokens[appId]

  // Auto-login using SSO when component mounts
  useEffect(() => {
    if (!loggedIn && hasAnyToken()) {
      handleSSOLogin()
    }
  }, [])

  const handleSSOLogin = async () => {
    setLoading(true)
    setError(null)

    try {
      await getTokenForApp(appId)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    logout()
  }

  if (loading) {
    return (
      <div className="loading">
        <h3>🔄 SSO in Progress...</h3>
        <p>Exchanging token from App A for App B token...</p>
      </div>
    )
  }

  if (loggedIn) {
    return (
      <div>
        <div className="app-header">
          <h2>📊 App B - Analytics</h2>
          <span className="status-badge logged-in">Logged In via SSO</span>
        </div>

        <div className="success">
          🎉 You were automatically logged in using Token Exchange (SSO)!
        </div>

        <div className="user-info">
          <h3>Welcome, {user?.name}!</h3>
          <p><strong>Email:</strong> {user?.email}</p>
          <p><strong>User ID:</strong> {user?.sub}</p>
        </div>

        <div className="sso-demo">
          <h3>✨ How SSO Worked:</h3>
          <ol style={{ marginLeft: '20px', marginTop: '10px' }}>
            <li>You logged into App A</li>
            <li>App B detected you have a token from App A</li>
            <li>App B used <strong>Token Exchange</strong> to get its own token</li>
            <li>No login required! 🎉</li>
          </ol>
        </div>

        <div className="button-group">
          <button className="btn btn-danger" onClick={handleLogout}>
            Logout from All Apps
          </button>
        </div>

        <details style={{ marginTop: '20px' }}>
          <summary style={{ cursor: 'pointer', fontWeight: 'bold' }}>
            🔍 View Access Token (Different from App A!)
          </summary>
          <div className="token-display">
            {token?.access_token}
          </div>
        </details>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <div className="app-header">
          <h2>📊 App B - Analytics</h2>
          <span className="status-badge logged-out">Not Logged In</span>
        </div>

        <div className="error">{error}</div>

        <div className="sso-demo">
          <h3>⚠️ SSO Not Available</h3>
          <p>Please login to App A first to enable SSO.</p>
        </div>

        <button className="btn btn-secondary" onClick={handleSSOLogin}>
          Try SSO Login Again
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="app-header">
        <h2>📊 App B - Analytics</h2>
        <span className="status-badge logged-out">Not Logged In</span>
      </div>

      <div className="sso-demo">
        <h3>🔐 SSO Available!</h3>
        <p>You have a token from another app. Click below to automatically login using Token Exchange.</p>
      </div>

      <button className="btn btn-secondary" onClick={handleSSOLogin} style={{ width: '100%' }}>
        🚀 Login with SSO (Token Exchange)
      </button>

      <div style={{ marginTop: '20px', padding: '15px', background: '#f5f5f5', borderRadius: '4px' }}>
        <p style={{ fontSize: '14px', color: '#666' }}>
          <strong>How it works:</strong><br/>
          1. You logged into App A<br/>
          2. App B will exchange App A's token for its own token<br/>
          3. No password required!
        </p>
      </div>
    </div>
  )
}

export default AppB
