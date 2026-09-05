import { useState, useEffect } from 'react';

const BACKEND_URL = 'http://localhost:3000/api/auth/token';
const REDIRECT_URI = 'http://localhost:5173';
const RC_CLIENT_ID = 'EMGGeo9Ve3scRNKDgUhN02Su0hx7fMRkIOcQLp42JgA';
const RC_AUTH_URL = 'https://www.recurse.com/oauth/authorize';

type TokenResponse = {
  access_token: string;
  token_type: string;
  scope?: string;
}

function generateCodeVerifier(): string {
  const array = new Uint32Array(56);
  window.crypto.getRandomValues(array);
  return Array.from(array, (dec) => ('0' + dec.toString(16)).substr(-2)).join('');
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const hash = await window.crypto.subtle.digest('SHA-256', data);

  return btoa(String.fromCharCode(...new Uint8Array(hash)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

export const useAuth = () => {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async () => {
    const verifier = generateCodeVerifier();
    const challenge = await generateCodeChallenge(verifier);

    sessionStorage.setItem('pkce_code_verifier', verifier);

    const params = new URLSearchParams({
      client_id: RC_CLIENT_ID,
      redirect_uri: REDIRECT_URI,
      response_type: 'code',
      code_challenge: challenge,
      code_challenge_method: 'S256'
    })

    window.location.href = `${RC_AUTH_URL}?${params.toString()}`;
  };

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');

    if (!code) {
      return;
    }

    const exchangeCodeForToken = async (code: string) => {
      setLoading(true);
      setError(null);

      const codeVerifier = sessionStorage.getItem('pkce_code_verifier') || '';

      try {
        const response = await fetch(BACKEND_URL, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            code,
            code_verifier: codeVerifier,
            redirect_uri: REDIRECT_URI
          })
        });

        const data = await response.json();
        if (!response.ok) {
          throw new Error(data.error || 'Failed to exchange token.');
        }

        const tokenData = data as TokenResponse;
        setToken(tokenData.access_token);

        // Clear URL query params
        window.history.replaceState({}, document.title, window.location.pathname);
      } catch (err) {
        setError(err.message || 'An error occurred during authentication.');
      } finally {
        setLoading(false);
      }
    };

    window.history.replaceState({}, document.title, window.location.pathname);
    exchangeCodeForToken(code);
  }, []);

  return {
    token, loading, error, handleLogin
  };
};
