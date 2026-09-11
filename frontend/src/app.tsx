import { useState, useEffect, useRef } from 'react';
import { RecurseLogin } from './auth/recurse';
import { DebugLogin } from './auth/debug';
import { Terminal } from './terminal/terminal';

export const App = () => {
  const terminalPrompt = ">";

  // Auth state
  const [token, setToken] = useState<string | null>(null);
  const useDebugAuth = import.meta.env.VITE_ENABLE_DEBUG_AUTH === 'true';

  // Terminal state
  const [command, setCommand] = useState('');
  const [lines, setLines] = useState<string[]>([]);

  // Web socket
  const webSocketRef = useRef(null);

  const terminalWriteLine = (line: string) => setLines((previous) => [...previous, line]);
  const onSubmit = (command: string) => {
    if (webSocketRef.current === null) {
      return;
    }

    terminalWriteLine(`\n${terminalPrompt} ${command}`);
    webSocketRef.current.send(command.trim());
  };

  // Init web socket
  useEffect(() => {
    if (token === null) {
      return;
    }

    // Create web socket
    // TODO: configure for prod
    webSocketRef.current = new WebSocket(`ws://${window.location.hostname}:5173/api/websocket?token=${token}`);
    console.log('Created web socket.');

    // Web socket open listener
    webSocketRef.current.addEventListener('open', () => {
      console.log('Web socket connected.');
    });

    // Web socket message listener
    webSocketRef.current.addEventListener('message', (messageEvent) => {
      terminalWriteLine(messageEvent.data);
    });

    // Web socket close listener
    webSocketRef.current.addEventListener('close', () => {
      terminalWriteLine('The server has disconnected.');
    });

    return () => {
      webSocketRef.current.removeEventListener('open');
      webSocketRef.current.removeEventListener('message');
      webSocketRef.current.removeEventListener('close');
      webSocketRef.current.close();
    };
  }, [token]);

  return (
    <div>
      <h1>RC Disco MUD!</h1>
      { useDebugAuth && <DebugLogin token={token} setToken={setToken} /> }
      { !useDebugAuth && <RecurseLogin token={token} setToken={setToken} /> }
      <Terminal prompt={terminalPrompt} lines={lines} command={command} setCommand={setCommand} onSubmit={onSubmit} />
    </div>
  );
};
