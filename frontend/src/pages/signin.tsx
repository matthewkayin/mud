import { useState, type FormEvent } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';

export function SignInPage() {
  const { user, loading, debugAuth, signIn, debugSignIn } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  if (loading) {
    return null;
  }
  if (user) {
    return <Navigate to="/game" replace />;
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setPending(true);
    try {
      if (debugAuth) {
        await debugSignIn(username);
      } else {
        await signIn(username, password);
      }
      navigate('/game');
    } catch (submitError) {
      setError((submitError as Error).message);
    } finally {
      setPending(false);
    }
  }

  return (
    <article>
      <h1>Sign In</h1>
      <form className="form" onSubmit={handleSubmit}>
        <fieldset>
          <legend>Your credentials</legend>
          {debugAuth && <p className="notice">Debug sign-in: no password needed.</p>}
          {error && <p className="notice">{error}</p>}
          <label htmlFor="username">Username</label>
          <input
            id="username"
            autoComplete="username"
            required
            value={username}
            onChange={(event) => setUsername(event.target.value)}
          />
          {!debugAuth && (
            <>
              <label htmlFor="password">Password</label>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </>
          )}
          <div className="form-actions">
            <button type="submit" disabled={pending}>{pending ? 'Signing in…' : 'Sign in'}</button>
          </div>
        </fieldset>
      </form>
      <p>No account yet? <Link to="/create-account">Create one</Link>.</p>
    </article>
  );
}
