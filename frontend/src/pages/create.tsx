import { useState, type FormEvent } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';

export function CreateAccountPage() {
  const { user, loading, createAccount } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
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
    if (password !== confirmPassword) {
      setError('The passwords do not match.');
      return;
    }
    setError(null);
    setPending(true);
    try {
      await createAccount(username, password);
      navigate('/game');
    } catch (submitError) {
      setError((submitError as Error).message);
    } finally {
      setPending(false);
    }
  }

  return (
    <article>
      <h1>Create an Account</h1>
      <form className="form" onSubmit={handleSubmit}>
        <fieldset>
          <legend>Choose your credentials</legend>
          {error && <p className="notice">{error}</p>}
          <label htmlFor="username">Username</label>
          <input
            id="username"
            autoComplete="username"
            required
            value={username}
            onChange={(event) => setUsername(event.target.value)}
          />
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            required
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
          <label htmlFor="confirm-password">Confirm password</label>
          <input
            id="confirm-password"
            type="password"
            autoComplete="new-password"
            required
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
          />
          <div className="form-actions">
            <button type="submit" disabled={pending}>{pending ? 'Creating…' : 'Create account'}</button>
          </div>
        </fieldset>
      </form>
      <p>Already have an account? <Link to="/signin">Sign in</Link>.</p>
    </article>
  );
}
