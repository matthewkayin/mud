import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { MudTerminal } from './terminal';

const socket = new WebSocket("ws://localhost:8080");
socket.addEventListener("open", () => {
  console.log("Connected to server.");
  setInterval(() => socket.send("hello it's me."), 1000);
});

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <MudTerminal/>
  </StrictMode>,
)
