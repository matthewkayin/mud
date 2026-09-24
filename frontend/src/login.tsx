import { useState } from 'react';

export const LoginPage = () => {
  const debugAuth = import.meta.env.VITE_ENABLE_DEBUG_AUTH === 'true';
  const [debugUser, setDebugUser] = useState('');

  const handleLogin = (event) => {
    event.preventDefault();

    if (debugAuth) {
      const parsedUser = Number.parseInt(debugUser);
      if (Number.isNaN(parsedUser)) {
        return;
      }
      if (parsedUser < 0) {
        return;
      }

      window.location.href = `http://${window.location.hostname}:5173/api/auth/debug?user=${debugUser}`;
      return;
    }

    window.location.href = `http://${window.location.hostname}:5173/api/auth/login?hostname=${window.location.hostname}`;
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
