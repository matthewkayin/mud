import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './app.tsx';

const authConfig: TAuthConfig = {
  clientId: '3zRJK2nQe62dlv9hu0urQABEvZm4jyYb3LdOSQ579ZI',
  authorizationEndpoint: 'https://recurse.com/oauth/authorize',
  tokenEndpoint: 'https://recurse.com/oauth/token',
  redirectUri: 'http://localhost:5173/',
  extraAuthParameters: {
      'Access-Control-Allow-Origin': 'http://localhost:5173'
  }
};

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <div style={{
      minHeight: '100vh',
    }}>
      <App />
    </div>
  </StrictMode>,
)
