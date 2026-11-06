import { Link } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';

export function HomePage() {
  const { user } = useAuth();

  return (
    <article>
      <h1>Welcome, Traveler</h1>
      <p>
        <span className="drop-cap">C</span>astle Recurse is a multiplayer online text adventure
        made by and for members of the Recurse Center community. Wander the halls of Castle Recurse and the wilds of Brooklyn beyond, but don't go alone!
        Every other player is in the same world in real time, so you may meet them on the stairs,
        trade with them, or brave the world's dangers together.
      </p>

      <hr />

      <p className="center">
        {user ? (
          <Link to="/game" className="big-link">Enter the castle &raquo;</Link>
        ) : (
          <>
            <Link to="/signin" className="big-link">Sign in to play &raquo;</Link>
            <br />
            <small>New here? <Link to="/create-account">Create an account</Link>.</small>
          </>
        )}
      </p>
    </article>
  );
}
