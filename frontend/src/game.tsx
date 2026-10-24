import { useState, useEffect, useRef } from 'react';
import './game.css';
import { Terminal } from './terminal/terminal';
import { StatusBar } from './status_bar';
import { getPlayerStatus, parseSocketMessage, type PlayerStatus } from './socket_message';

const terminalPrompt = ">";

// The server sends a heartbeat every 3 seconds, so this allows for a few missed heartbeats
const HEARTBEAT_TIMEOUT_MS = 10000;

export const GamePage = () => {
  // Terminal state
  const [command, setCommand] = useState('');
  const [lines, setLines] = useState<string[]>([]);
  const [playerStatus, setPlayerStatus] = useState<PlayerStatus | null>(null);

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

    // If nothing arrives from the server for too long, the connection is treated as dead
    let heartbeatTimeout: ReturnType<typeof setTimeout> | undefined;
    const resetHeartbeatTimeout = () => {
      clearTimeout(heartbeatTimeout);
      heartbeatTimeout = setTimeout(() => {
        console.log('Web socket heartbeat timed out.');
        handleDisconnect();
      }, HEARTBEAT_TIMEOUT_MS);
    };

    // Web socket open listener
    const onOpen = () => {
      console.log('Web socket connected.');
      resetHeartbeatTimeout();
    };

    // Web socket message listener
    const onMessage = (messageEvent: MessageEvent) => {
      const socketMessage = parseSocketMessage(messageEvent.data);
      if (socketMessage === null) {
        console.warn('Received an invalid message from the server.', messageEvent.data);
        return;
      }

      resetHeartbeatTimeout();
      if (socketMessage.message !== null) {
        terminalWriteLine(socketMessage.message);
      }

      // A heartbeat without a status means the player isn't in-game
      const status = getPlayerStatus(socketMessage);
      if (status !== null || socketMessage.message === null) {
        setPlayerStatus(status);
      }
    };

    // Web socket close listener
    const onClose = () => {
      handleDisconnect();
    };

    const removeListeners = () => {
      webSocket.removeEventListener('open', onOpen);
      webSocket.removeEventListener('message', onMessage);
      webSocket.removeEventListener('close', onClose);
    };

    // Called both when the socket closes and when the heartbeat times out,
    // since closing a dead socket can take a long time to fire the close event
    let isDisconnected = false;
    const handleDisconnect = () => {
      if (isDisconnected) {
        return;
      }
      isDisconnected = true;

      clearTimeout(heartbeatTimeout);
      removeListeners();
      webSocket.close();
      webSocketRef.current = null;
      setPlayerStatus(null);
      terminalWriteLine('The server has disconnected.');
    };

    webSocket.addEventListener('open', onOpen);
    webSocket.addEventListener('message', onMessage);
    webSocket.addEventListener('close', onClose);

    return () => {
      clearTimeout(heartbeatTimeout);
      removeListeners();
      webSocket.close();
    };
  }, []);

  return (
    <div>
      <h1>Castle Recurse</h1>
      <Terminal prompt={terminalPrompt} lines={lines} command={command} setCommand={setCommand} onSubmit={onSubmit} />
      <div className="game-hud">
        <StatusBar status={playerStatus} />
      </div>
    </div>
  );
};
