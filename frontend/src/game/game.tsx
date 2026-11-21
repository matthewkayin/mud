import { useState, useEffect, useRef } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../auth/auth_context.tsx';
import { getDebugUsername } from '../auth/auth_api.ts';
import { Terminal } from './terminal.tsx';
import { StatusBar } from './status_bar.tsx';
import { Minimap } from './minimap.tsx';
import { View } from './view.tsx';
import { VIEW_MODES, parseViewMode, type ViewMode } from './view_mode.ts';
import { type EquipmentSlot, type Inventory, type PlayerStatus, type RoomMob, type CharacterSheet, type SocketMessage } from './socket_message';

const TERMINAL_PROMPT = ">";
const HEARTBEAT_TIMEOUT_MS = 10000;
const TERMINAL_MAX_LINE_COUNT = 1024;
const TERMINAL_MAX_HISTORY = 32;
const CHAT_MAX_LINE_COUNT = 1024;
// A chat message starts with a bracketed tag, such as "[Chat]", optionally wrapped in a color tag
const CHAT_MESSAGE_PATTERN = /^(<[a-z]*>)?\[[^\]]+\]/;

// Appends a line, dropping the oldest line once there are maxLineCount lines
const appendLine = (lines: string[], line: string, maxLineCount: number) => {
  const startIndex = lines.length >= maxLineCount ? lines.length - maxLineCount + 1 : 0;
  return [...lines.slice(startIndex), line];
};

export function GamePage() {
  // Auth
  const { user, loading } = useAuth();

  // Terminal state
  const [command, setCommand] = useState('');
  const [lines, setLines] = useState<string[]>([]);
  const [history, setHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);

  // Status
  const [playerStatus, setPlayerStatus] = useState<PlayerStatus | null>(null);
  const [minimap, setMinimap] = useState<number[] | null>(null);

  // View panel
  const [viewMode, setViewMode] = useState<ViewMode>('help');
  const [chatLines, setChatLines] = useState<string[]>([]);
  const [inventory, setInventory] = useState<Inventory | null>(null);
  const [equipment, setEquipment] = useState<EquipmentSlot[] | null>(null);
  const [roomMobs, setRoomMobs] = useState<RoomMob[] | null>(null);
  const [characterSheet, setCharacterSheet] = useState<CharacterSheet | null>(null);

  // The player is in the world while the server is sending them a status
  const isInWorld = playerStatus !== null;

  // Websocket
  const websocketRef = useRef<WebSocket | null>(null);

  const terminalWriteLine = (line: string) => setLines((previous) => appendLine(previous, line, TERMINAL_MAX_LINE_COUNT));

  // The client handles the view command itself, since it already has the lines for every mode
  const handleViewCommand = (args: string[]) => {
    if (!isInWorld) {
      terminalWriteLine("You can only use 'view' while you are in the world.");
      return;
    }

    const mode = parseViewMode(args.length === 1 ? args[0].toLowerCase() : '');
    if (mode === null) {
      terminalWriteLine(`Usage: view <${VIEW_MODES.join('|')}>`);
      return;
    }

    setViewMode(mode);
  };
  const onSubmit = (command: string) => {
    if (websocketRef.current === null) {
      return;
    }

    terminalWriteLine(`\n${TERMINAL_PROMPT} ${command}`)
    setHistory((previous) => {
      const startIndex = previous.length == TERMINAL_MAX_HISTORY ? 1 : 0;
      const previousHistory = previous.slice(startIndex);
      return [command, ...previousHistory];
    });
    setHistoryIndex(-1);

    const words = command.trim().split(/\s+/);
    if (words[0].toLowerCase() === 'view') {
      handleViewCommand(words.slice(1));
      return;
    }

    websocketRef.current.send(command.trim());
  }

  // Init websocket
  useEffect(() => {
    // Create websocket timeout
    const websocketProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    // A debug sign-in has no session cookie, so the username goes in the URL
    const debugUsername = getDebugUsername();
    const websocketQuery = debugUsername === null ? '' : `?debugUser=${encodeURIComponent(debugUsername)}`;
    const websocket = new WebSocket(`${websocketProtocol}://${window.location.host}/api/websocket${websocketQuery}`);
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

        const message = socketMessage.message;
        if (CHAT_MESSAGE_PATTERN.test(message)) {
          setChatLines((previous) => appendLine(previous, message, CHAT_MAX_LINE_COUNT));
        }
      }

      // Get status from socketMessage
      if (socketMessage.status) {
        setPlayerStatus(socketMessage.status);
      } else if (socketMessage.message === null) {
        // If we receive no status and message is null, then this is a status-less heartbeat,
        // which means the player is receiving a heartbeat but they aren't in game, so clear
        // their status, minimap, inventory, equipment, room mobs and character sheet
        setPlayerStatus(null);
        setMinimap(null);
        setInventory(null);
        setEquipment(null);
        setRoomMobs(null);
        setCharacterSheet(null);
      }

      // Get minimap from socketMessage
      // The server only sends the minimap when it changes
      if (socketMessage.map) {
        setMinimap(socketMessage.map);
      }

      if (socketMessage.inv) {
        setInventory(socketMessage.inv);
      }
      if (socketMessage.equip) {
        setEquipment(socketMessage.equip);
      }
      // The server only sends the room mobs when they change
      if (socketMessage.mobs) {
        setRoomMobs(socketMessage.mobs);
      }
      // The server only sends the character sheet when it changes
      if (socketMessage.sheet) {
        setCharacterSheet(socketMessage.sheet);
      }
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
      setPlayerStatus(null);
      setMinimap(null);
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
      <div className="game-main">
        <Terminal
          prompt={TERMINAL_PROMPT}
          lines={lines}
          command={command}
          setCommand={setCommand}
          history={history}
          historyIndex={historyIndex}
          setHistoryIndex={setHistoryIndex}
          pendingAction={playerStatus?.action ?? false}
          onSubmit={onSubmit}
        />
        <StatusBar status={playerStatus} />
      </div>
      <div className="game-sidebar">
        <Minimap map={minimap} />
        <View
          mode={viewMode}
          isInWorld={isInWorld}
          chatLines={chatLines}
          inventory={inventory}
          equipment={equipment}
          roomMobs={roomMobs}
          characterSheet={characterSheet}
        />
      </div>
    </article>
  );
}
