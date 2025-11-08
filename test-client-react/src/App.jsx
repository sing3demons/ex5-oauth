import { useState } from 'react'
import AppA from './components/AppA'
import AppB from './components/AppB'
import AppC from './components/AppC'
import { SSOProvider } from './context/SSOContext'

function App() {
  const [activeApp, setActiveApp] = useState('app-a')

  return (
    <SSOProvider>
      <div className="container">
        <h1 style={{ textAlign: 'center', marginBottom: '30px', color: '#333' }}>
          🔐 OAuth2 SSO Test Client
        </h1>
        
        <div className="sso-demo">
          <h3>🎯 SSO Demo Instructions:</h3>
          <ol style={{ marginLeft: '20px', marginTop: '10px' }}>
            <li>Login to <strong>App A</strong> first</li>
            <li>Switch to <strong>App B</strong> - it will automatically get a token using Token Exchange (SSO!)</li>
            <li>Switch to <strong>App C</strong> - same automatic SSO!</li>
            <li>Try logging out from any app - all apps will be logged out</li>
          </ol>
        </div>

        <div className="app-selector">
          <button
            className={`app-button ${activeApp === 'app-a' ? 'active' : ''}`}
            onClick={() => setActiveApp('app-a')}
          >
            📱 App A<br/>
            <small>E-commerce</small>
          </button>
          <button
            className={`app-button ${activeApp === 'app-b' ? 'active' : ''}`}
            onClick={() => setActiveApp('app-b')}
          >
            📊 App B<br/>
            <small>Analytics</small>
          </button>
          <button
            className={`app-button ${activeApp === 'app-c' ? 'active' : ''}`}
            onClick={() => setActiveApp('app-c')}
          >
            💬 App C<br/>
            <small>Chat</small>
          </button>
        </div>

        <div className="app-content">
          {activeApp === 'app-a' && <AppA />}
          {activeApp === 'app-b' && <AppB />}
          {activeApp === 'app-c' && <AppC />}
        </div>
      </div>
    </SSOProvider>
  )
}

export default App
