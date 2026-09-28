import { StrictMode } from 'react';
import { CssBaseline } from '@mui/material';
import { createRoot } from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { LoginPage } from './login.tsx';
import { GamePage } from './game.tsx';
import { Editor } from './editor/editor.tsx';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <CssBaseline/>
    <BrowserRouter>
      <div style={{
        minHeight: '100vh',
      }}>
        <Routes>
          <Route path="/" element={<LoginPage/>}/>
          <Route path="/game" element={<GamePage/>}/>
          <Route path="/editor" element={<Editor/>}/>
        </Routes>
      </div>
    </BrowserRouter>
  </StrictMode>,
)
