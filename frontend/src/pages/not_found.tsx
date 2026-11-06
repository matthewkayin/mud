import { Link } from 'react-router-dom';

export function NotFoundPage() {
  return (
    <article>
      <h1>You seem lost, friend.</h1>
      <p>This passage leads nowhere. <Link to="/">Return home</Link>.</p>
    </article>
  );
}
