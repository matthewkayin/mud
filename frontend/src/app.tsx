import { useEffect, useRef } from 'react';
import { useAuth } from './auth';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

const TERMINAL_PROMPT = '>';

export const App = () => {
  // Auth
  const { token, loading, error, handleLogin } = useAuth();

  // Terminal state
  const terminalContainerRef = useRef(null);
  const terminalInstanceRef = useRef(null);
  const terminalCommandBufferRef = useRef('');

  // Web socket
  const webSocketRef = useRef(null);

  // Define draw bottom prompt function
  const terminalDrawBottomPrompt = () => {
    if (!terminalInstanceRef.current) {
      console.log('Warn: called terminalDrawBottomPrompt but terminalInstanceRef is empty.');
      return;
    }

    // Move cursor to the bottom row, first column
    terminalInstanceRef.current.write(`\x1b[${terminalInstanceRef.current.rows};1H`);

    // Clear the entire bottom line
    terminalInstanceRef.current.write('\x1b[2K');

    // Write prompt and current user input
    terminalInstanceRef.current.write(`${TERMINAL_PROMPT} ${terminalCommandBufferRef.current}`);
  };

  const terminalWriteLine = (message: string) => {
    if (!terminalInstanceRef.current) {
      console.log('Warn: called terminalWriteLine but terminalInstanceRef is empty.');
      return;
    }

    terminalInstanceRef.current.write(`\x1b[${terminalInstanceRef.current.rows - 1};1H${message}\r\n`);
    terminalDrawBottomPrompt();
  };

  // Init terminal
  useEffect(() => {
    if (!terminalContainerRef.current) {
      return;
    }

    const terminal = new Terminal({
      cursorBlink: true,
      theme: {
        background: '#1e1e1e',
        foreground: '#ffffff'
      }
    });
    terminal.open(terminalContainerRef.current);
    terminalInstanceRef.current = terminal;
    console.log('Set term instance.');

    // Add fit add-on
    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    fitAddon.fit();

    // On resize handler
    const onResize = () => {
      fitAddon.fit();
    };
    window.addEventListener('resize', onResize);

    // Set the scrolling region to exclude the bottom row
    const promptRow = terminal.rows - 1;
    terminal.write(`\x1b[1;${promptRow}r`);

    // Define data listener
    const dataListener = terminal.onData((data) => {
      // Submit
      if (data === '\r') {
        const command = terminalCommandBufferRef.current.trim();
        terminalCommandBufferRef.current = '';
        terminal.write('\r\n');

        if (webSocketRef.current && command) {
          webSocketRef.current.send(command);
        } else {
          console.log('Warning: terminal on submit not handled.');
        }

        terminalDrawBottomPrompt();

        return;
      }

      // Backspace
      if (data === '\u007F') {
        // Writes three characters
        // One moves the cursor back, one inserts a space, the other moves the cursor back again
        terminal.write('\b \b');
        return;
      }

      // Append visible characters to command
      if (data >= ' ' && data <= '~') {
        terminal.write(data);
        terminalCommandBufferRef.current += data;
      }
    });

    terminalDrawBottomPrompt();

    // Cleanup - called on component unmount
    return () => {
      window.removeEventListener('resize', onResize);
      dataListener.dispose();
      terminal.dispose();
    };
  }, []);

  // Init web socket
  useEffect(() => {
    if (token === null) {
      return;
    }

    webSocketRef.current = new WebSocket('ws://localhost:3000/ws-telnet', [token]);
    console.log('Created web socket.');
    webSocketRef.current.addEventListener('open', () => {
      console.log('Web socket connected.');

      if (!terminalInstanceRef.current) {
        console.log('Term instance is empty, exiting.');
        return;
      }

      terminalDrawBottomPrompt();
    });
    webSocketRef.current.addEventListener('message', (messageEvent) => {
      console.log('Web socket received message: ', messageEvent);

      if (!terminalInstanceRef.current) {
        console.log('Term instance is empty, exiting.');
        return;
      }

      terminalWriteLine(messageEvent.data);
    });
    webSocketRef.current.addEventListener('close', () => {
      if (!terminalInstanceRef.current) {
        return;
      }

      terminalWriteLine('The server has disconnected.');
    });

    return () => {
      webSocketRef.current.removeEventListener('open');
      webSocketRef.current.removeEventListener('message');
      webSocketRef.current.removeEventListener('close');
    };
  }, [token]);

  return (
    <div>
      <h1>RC Disco MUD!</h1>
      {loading && <p>Authorizing...</p>}
      {error && <p>Error: {error}</p>}
      {(!token && !loading) && <button onClick={handleLogin}>Login with Recurse</button>}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
      }}>
        <div
          ref={terminalContainerRef}
          style={{
            width: '100vw',
            height: '600px',
            backgroundColor: '#1e1e1e',
            padding: '10px',
            borderRadius: '4px',
          }}
        />
      </div>
    </div>
  );
};
