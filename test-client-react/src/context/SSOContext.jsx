import { createContext, useContext, useState, useEffect } from 'react'

const SSOContext = createContext()

const OAUTH_SERVER = 'http://localhost:8080'

// Client configurations
const CLIENTS = {
  'app-a': {
    client_id: 'app-a-client-id',
    client_secret: 'app-a-secret',
    redirect_uri: 'http://localhost:3000/callback',
    name: 'App A (E-commerce)'
  },
  'app-b': {
    client_id: 'app-b-client-id',
    client_secret: 'app-b-secret',
    redirect_uri: 'http://localhost:3000/callback',
    name: 'App B (Analytics)'
  },
  'app-c': {
    client_id: 'app-c-client-id',
    client_secret: 'app-c-secret',
    redirect_uri: 'http://localhost:3000/callback',
    name: 'App C (Chat)'
  }
}

export function SSOProvider({ children }) {
  const [tokens, setTokens] = useState(() => {
    // Load tokens from localStorage
    const stored = localStorage.getItem('sso_tokens')
    return stored ? JSON.parse(stored) : {}
  })

  const [users, setUsers] = useState(() => {
    const stored = localStorage.getItem('sso_users')
    return stored ? JSON.parse(stored) : {}
  })

  // Save tokens to localStorage whenever they change
  useEffect(() => {
    localStorage.setItem('sso_tokens', JSON.stringify(tokens))
  }, [tokens])

  useEffect(() => {
    localStorage.setItem('sso_users', JSON.stringify(users))
  }, [users])

  // Login to an app (simplified - direct token request)
  const login = async (appId, email, password) => {
    try {
      const client = CLIENTS[appId]
      
      // For demo purposes, we'll use a simplified flow:
      // Direct login to get tokens (bypassing full OAuth flow)
      
      // Step 1: Login and get tokens directly
      // Note: This is a simplified demo flow. In production, you'd use full OAuth flow.
      const loginResponse = await fetch(`${OAUTH_SERVER}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({
          email,
          password
        })
      })

      if (!loginResponse.ok) {
        const error = await loginResponse.json()
        throw new Error(error.error_description || 'Login failed')
      }

      const loginData = await loginResponse.json()
      
      // loginData contains access_token, refresh_token, etc.
      const tokenData = {
        access_token: loginData.access_token,
        token_type: loginData.token_type || 'Bearer',
        expires_in: loginData.expires_in,
        refresh_token: loginData.refresh_token,
        scope: loginData.scope || 'openid profile email'
      }

      // Get user info
      const userInfo = await getUserInfo(tokenData.access_token)

      // Store tokens and user info
      setTokens(prev => ({
        ...prev,
        [appId]: tokenData
      }))

      setUsers(prev => ({
        ...prev,
        [appId]: userInfo
      }))

      return { tokens: tokenData, user: userInfo }
    } catch (error) {
      console.error('Login error:', error)
      throw error
    }
  }

  // Get token for another app using Token Exchange (SSO!)
  const getTokenForApp = async (targetAppId) => {
    try {
      // Check if we already have a token for this app
      if (tokens[targetAppId]) {
        const token = tokens[targetAppId]
        // Check if token is still valid
        if (!isTokenExpired(token.access_token)) {
          return token
        }
      }

      // Find any valid token from other apps
      const sourceToken = findValidToken()
      if (!sourceToken) {
        throw new Error('No valid token available. Please login first.')
      }

      const targetClient = CLIENTS[targetAppId]

      // Exchange token using Token Exchange grant
      const response = await fetch(`${OAUTH_SERVER}/oauth/token`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: new URLSearchParams({
          grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
          subject_token: sourceToken,
          subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
          requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
          client_id: targetClient.client_id,
          client_secret: targetClient.client_secret,
          scope: 'openid profile email'
        })
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error_description || 'Token exchange failed')
      }

      const tokenData = await response.json()

      // Get user info
      const userInfo = await getUserInfo(tokenData.access_token)

      // Store tokens and user info
      setTokens(prev => ({
        ...prev,
        [targetAppId]: tokenData
      }))

      setUsers(prev => ({
        ...prev,
        [targetAppId]: userInfo
      }))

      return tokenData
    } catch (error) {
      console.error('Token exchange error:', error)
      throw error
    }
  }

  // Get user info from UserInfo endpoint
  const getUserInfo = async (accessToken) => {
    const response = await fetch(`${OAUTH_SERVER}/oauth/userinfo`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    })

    if (!response.ok) {
      throw new Error('Failed to get user info')
    }

    return await response.json()
  }

  // Find any valid token
  const findValidToken = () => {
    for (const appId in tokens) {
      const token = tokens[appId].access_token
      if (!isTokenExpired(token)) {
        return token
      }
    }
    return null
  }

  // Check if token is expired
  const isTokenExpired = (token) => {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      return payload.exp * 1000 < Date.now()
    } catch {
      return true
    }
  }

  // Logout from all apps
  const logout = () => {
    setTokens({})
    setUsers({})
    localStorage.removeItem('sso_tokens')
    localStorage.removeItem('sso_users')
  }

  // Logout from specific app
  const logoutApp = (appId) => {
    setTokens(prev => {
      const newTokens = { ...prev }
      delete newTokens[appId]
      return newTokens
    })

    setUsers(prev => {
      const newUsers = { ...prev }
      delete newUsers[appId]
      return newUsers
    })
  }

  const value = {
    tokens,
    users,
    login,
    getTokenForApp,
    logout,
    logoutApp,
    isLoggedIn: (appId) => !!tokens[appId] && !isTokenExpired(tokens[appId].access_token),
    hasAnyToken: () => Object.keys(tokens).length > 0
  }

  return <SSOContext.Provider value={value}>{children}</SSOContext.Provider>
}

export function useSSO() {
  const context = useContext(SSOContext)
  if (!context) {
    throw new Error('useSSO must be used within SSOProvider')
  }
  return context
}
