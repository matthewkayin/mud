import { useState } from 'react';

export const LoginPage = () => {
  const debugAuth = import.meta.env.VITE_ENABLE_DEBUG_AUTH;
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

    console.log('Navigate! Window location origin is :', window.location.origin);
    window.location.href = `http://${window.location.hostname}:5173/api/auth/login?hostname=${window.location.origin}`;
  };

  return (
    <div>
      <h1>RC Disco MUD!</h1>
      <form onSubmit = {handleLogin}>
        <div>
          { debugAuth && <input value={debugUser} onChange={(event) => setDebugUser(event.target.value)}></input> }
          <button onClick={handleLogin}>Login with Recurse</button>
        </div>
    </form>
    </div>
  )
};
