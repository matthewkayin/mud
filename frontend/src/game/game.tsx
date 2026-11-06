import { useState, useEffect, useRef } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';
import { Terminal } from './terminal.tsx';
import { StatusBar } from './status_bar.tsx';
import { type PlayerStatus, type SocketMessage } from './socket_message';

const TERMINAL_PROMPT = ">";
const HEARTBEAT_TIMEOUT_MS = 10000;

export function GamePage() {
  // Auth
  const { user, loading } = useAuth();

  // Terminal state
  const [command, setCommand] = useState('');
  const [lines, setLines] = useState<string[]>([]);

  // Status
  const [playerStatus, setPlayerStatus] = useState<PlayerStatus | null>(null);

  // Websocket
  const websocketRef = useRef<WebSocket | null>(null);

  const terminalWriteLine = (line: string) => setLines((previous) => [...previous, line]);
  const onSubmit = (command: string) => {
    if (websocketRef.current === null) {
      return;
    }

    terminalWriteLine(`\n${TERMINAL_PROMPT} ${command}`)
    websocketRef.current.send(command.trim());
  }

  // Init websocket
  useEffect(() => {
    // Create websocket timeout
    const websocketProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const websocket = new WebSocket(`${websocketProtocol}://${window.location.host}/api/websocket`);
    websocketRef.current = websocket;
    console.log('Created web socket.');

    // Check for heartbeat timeout
    // If nothing arrives from the server for too long, the connection is treated as dead
    let heartbeatTimeout: ReturnType<typeof setTimeout> | undefined;
    const resetHeartbeatTimeout = () => {
      clearTimeout(heartbeatTimeout);
      heartbeatTimeout = setTimeout(() => {
        console.log('Web socket heartbeat timed out.');
      }, HEARTBEAT_TIMEOUT_MS);
    }

    // Web socket open listener
    const onOpen = () => {
      console.log('Web socket connected.');
      resetHeartbeatTimeout();
    };

    // Web socket message listener
    const onMessage = (messageEvent: MessageEvent) => {
      const socketMessage = JSON.parse(messageEvent.data) as SocketMessage;
      if (socketMessage === null) {
        console.warn('Received an invalid message from the server.', messageEvent.data);
        return;
      }

      resetHeartbeatTimeout();

      // Get message from socketMessage
      if (socketMessage.message !== null) {
        terminalWriteLine(socketMessage.message);
      }

      console.log(socketMessage);

      // Get status from socketMessage
      if (socketMessage.status) {
        setPlayerStatus(socketMessage.status);
      } else if (socketMessage.message === null) {
        // If we receive no status and message is null, then this is a status-less heartbeat,
        // which means the player is receiving a heartbeat but they aren't in game, so clear
        // their status (which hides the status bar)
        setPlayerStatus(null);
      }

      // TODO: update player minimap
    };

    // Web socket close listener
    const onClose = () => {
      handleDisconnect();
    };

    const removeListeners = () => {
      websocket.removeEventListener('open', onOpen);
      websocket.removeEventListener('message', onMessage);
      websocket.removeEventListener('close', onClose);
    };

    // Handle disconnect is called both when the web socket closes and when the heartbeat times out
    let isDisconnected = false;
    const handleDisconnect = () => {
      if (isDisconnected) {
        return;
      }
      isDisconnected = true;

      clearTimeout(heartbeatTimeout);
      removeListeners();
      websocket.close();
      websocketRef.current = null;
      // TODO: clear player status and minimap
      terminalWriteLine('The server has disconnected.');
    };

    websocket.addEventListener('open', onOpen);
    websocket.addEventListener('message', onMessage);
    websocket.addEventListener('close', onClose);

    return () => {
      clearTimeout(heartbeatTimeout);
      removeListeners();
      websocket.close();
    };
  }, []);

  if (loading) {
    return null;
  }
  if (!user) {
    return <Navigate to="/signin" replace />;
  }

  return (
    <article className="game">
      <Terminal
        prompt={TERMINAL_PROMPT}
        lines={lines}
        command={command}
        setCommand={setCommand}
        pendingAction={false}
        onSubmit={onSubmit}
      />
      <StatusBar status={playerStatus} />
    </article>
  );
}
