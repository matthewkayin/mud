import { Link } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';

export function HomePage() {
  const { user } = useAuth();

  return (
    <article>
      <h1>Welcome, Traveler</h1>
      <p>
        <span className="drop-cap">C</span>astle Recurse is a <i>multi-user dungeon</i>: a text
        adventure that many people play at once. There are no pictures here, only words. The
        castle is described to you room by room, and you act upon it by typing commands such
        as <code>look</code>, <code>move north</code>, or <code>take lantern</code>.
      </p>
      <p>
        Wander its halls and the wilds beyond, fight what lurks in the dark, gather and craft,
        and grow stronger as you go. You will not be alone. Every other player is in the same
        world in real time, so you may meet them on the stairs, trade with them, or face the
        castle's dangers together.
      </p>
      <p>
        The game was made by and for the Recurse Center community, and it is still being built.
        Expect the castle to change beneath your feet.
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
