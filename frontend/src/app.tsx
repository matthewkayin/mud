import { useState, useEffect, useRef } from 'react';
import { useAuth } from './auth';
import { Terminal } from './terminal/terminal';

export const App = () => {
  // Auth state
  const { token, loading, error, handleLogin } = useAuth();

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

    webSocketRef.current.send(command.trim());
  };

  // Init web socket
  useEffect(() => {
    if (token === null) {
      return;
    }

    // Create web socket
    webSocketRef.current = new WebSocket(`ws://localhost:7272/api/websocket?token=${token}`);
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
      {loading && <p>Authorizing...</p>}
      {error && <p>Error: {error}</p>}
      {(!token && !loading) && <button onClick={handleLogin}>Login with Recurse</button>}
      <Terminal prompt=">" lines={lines} command={command} setCommand={setCommand} onSubmit={onSubmit} />
    </div>
  );
};
