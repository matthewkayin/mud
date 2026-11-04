import { Navigate } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';

export function GamePage() {
  const { user, loading } = useAuth();

  if (loading) {
    return null;
  }
  if (!user) {
    return <Navigate to="/signin" replace />;
  }

  return (
    <article>
      <h1>The Castle</h1>
      <div className="terminal-slot">
        {/* Terminal goes here */}
      </div>
    </article>
  );
}
