import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { LoginPage } from './login.tsx';
import { GamePage } from './game.tsx';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <div style={{
        minHeight: '100vh',
      }}>
        <Routes>
          <Route path="/" element={<LoginPage/ >}/>
          <Route path="/game" element={<GamePage/ >}/>
        </Routes>
      </div>
    </BrowserRouter>
  </StrictMode>,
)
