import { Route, Routes } from 'react-router-dom';
import { Layout } from './components/layout.tsx';
import { HomePage } from './pages/home.tsx';
import { SignInPage } from './pages/signin.tsx';
import { CreateAccountPage } from './pages/create.tsx';
import { GamePage } from './pages/game.tsx';
import { NotFoundPage } from './pages/not_found.tsx';

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/signin" element={<SignInPage />} />
        <Route path="/create-account" element={<CreateAccountPage />} />
        <Route path="/game" element={<GamePage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </Layout>
  );
}
