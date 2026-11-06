import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';

export function Nav() {
  const { user, loading, signOut } = useAuth();
  const navigate = useNavigate();
  const [signingOut, setSigningOut] = useState(false);

  async function handleSignOut() {
    setSigningOut(true);
    try {
      await signOut();
      navigate('/');
    } catch (error) {
      alert((error as Error).message);
    } finally {
      setSigningOut(false);
    }
  }

  return (
    <nav className="site-nav">
      <span className="nav-links">
        <Link to="/">Home</Link>
        <span className="nav-sep">|</span>
        <Link to="/game">Play</Link>
      </span>
      {!loading && (
        <span className="nav-links">
          {user ? (
            <>
              <span>Signed in as <b>{user.username}</b></span>
              <span className="nav-sep">|</span>
              <button type="button" className="link-button" onClick={handleSignOut} disabled={signingOut}>
                Log out
              </button>
            </>
          ) : (
            <>
              <Link to="/signin">Sign in</Link>
              <span className="nav-sep">|</span>
              <Link to="/create-account">Create account</Link>
            </>
          )}
        </span>
      )}
    </nav>
  );
}
