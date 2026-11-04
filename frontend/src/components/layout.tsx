import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { Nav } from './nav.tsx';

export function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="page">
      <header className="site-header">
        <Link to="/" className="site-title">Castle Recurse</Link>
        <p className="site-tagline">A multi-user dungeon for the Recurse Center</p>
      </header>
      <Nav />
      <main className="site-main">{children}</main>
      <footer className="site-footer">
        © 2026 Matthew Madden and Emanuel Stults, All Rights Reserved.
      </footer>
    </div>
  );
}
