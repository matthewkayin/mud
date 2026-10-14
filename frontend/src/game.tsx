import { useState, useEffect, useRef } from 'react';
import { Terminal } from './terminal/terminal';

const terminalPrompt = ">";

export const GamePage = () => {
  // Terminal state
  const [command, setCommand] = useState('');
  const [lines, setLines] = useState<string[]>([]);

  // Web socket
  const webSocketRef = useRef<WebSocket | null>(null);

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
    const webSocketProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const webSocket = new WebSocket(`${webSocketProtocol}://${window.location.host}/api/websocket`);
    webSocketRef.current = webSocket;
    console.log('Created web socket.');

    // TODO: set onerror listener and log to terminal on error (likely error is unauthorized)

    // Web socket open listener
    const onOpen = () => {
      console.log('Web socket connected.');
    };
    webSocket.addEventListener('open', onOpen);

    // Web socket message listener
    const onMessage = (messageEvent: MessageEvent) => {
      terminalWriteLine(messageEvent.data);
    };
    webSocket.addEventListener('message', onMessage);

    // Web socket close listener
    const onClose = () => {
      terminalWriteLine('The server has disconnected.');
    };
    webSocket.addEventListener('close', onClose);

    return () => {
      webSocket.removeEventListener('open', onOpen);
      webSocket.removeEventListener('message', onMessage);
      webSocket.removeEventListener('close', onClose);
      webSocket.close();
    };
  }, []);

  return (
    <div>
      <h1>Castle Recurse</h1>
      <Terminal prompt={terminalPrompt} lines={lines} command={command} setCommand={setCommand} onSubmit={onSubmit} />
    </div>
  );
};
