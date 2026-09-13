import { useState, useEffect, useRef } from 'react';
import { Terminal } from './terminal/terminal';

const terminalPrompt = ">";

export const GamePage = () => {
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
    // Create web socket
    // TODO: configure for prod
    webSocketRef.current = new WebSocket(`ws://${window.location.hostname}:5173/api/websocket`);
    console.log('Created web socket.');

    // TODO: set onerror listener and log to terminal on error (likely error is unauthorized)

    // Web socket open listener
    const onOpen = () => {
      console.log('Web socket connected.');
    };
    webSocketRef.current.addEventListener('open', onOpen);

    // Web socket message listener
    const onMessage = (messageEvent) => {
      terminalWriteLine(messageEvent.data);
    };
    webSocketRef.current.addEventListener('message', onMessage);

    // Web socket close listener
    const onClose = () => {
      terminalWriteLine('The server has disconnected.');
    };
    webSocketRef.current.addEventListener('close', onClose);

    return () => {
      webSocketRef.current.removeEventListener('open', onOpen);
      webSocketRef.current.removeEventListener('message', onMessage);
      webSocketRef.current.removeEventListener('close', onClose);
      webSocketRef.current.close();
    };
  }, []);

  return (
    <div>
      <h1>RC Disco MUD!</h1>
      <Terminal prompt={terminalPrompt} lines={lines} command={command} setCommand={setCommand} onSubmit={onSubmit} />
    </div>
  );
};
