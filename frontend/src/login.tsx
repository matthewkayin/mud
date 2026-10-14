import { useState, useEffect } from 'react';

export const LoginPage = () => {
  const [debugAuth, setDebugAuth] = useState(false);
  const [debugUser, setDebugUser] = useState('');

  // Ask the server whether debug auth is enabled
  useEffect(() => {
    fetch('/api/auth/config')
      .then((response) => response.json())
      .then((config) => setDebugAuth(config.debugAuth === true))
      .catch((error) => console.error('Failed to fetch auth config.', error));
  }, []);

  const handleLogin = (event: React.SyntheticEvent) => {
    event.preventDefault();

    if (debugAuth) {
      const parsedUser = Number.parseInt(debugUser);
      if (Number.isNaN(parsedUser)) {
        return;
      }
      if (parsedUser < 0) {
        return;
      }

      window.location.href = `/api/auth/debug?user=${debugUser}`;
      return;
    }

    window.location.href = '/api/auth/login';
  };

  return (
    <div>
      <h1>Castle Recurse</h1>
      <form onSubmit = {handleLogin}>
        <div>
          { debugAuth && <input value={debugUser} onChange={(event) => setDebugUser(event.target.value)}></input> }
          <button onClick={handleLogin}>Login with Recurse</button>
        </div>
    </form>
    </div>
  )
};
