import { useEffect, useRef } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

export const MudTerminal = () => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const xtermInstance = useRef<Terminal | null>(null);

  useEffect(() => {
    if (!terminalRef.current) {
      return;
    }

    const term = new Terminal({
      cursorBlink: true,
      theme: {
        background: '#1e1e1e',
        foreground: '#f8f8f2'
      }
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    term.open(terminalRef.current);
    fitAddon.fit();

    xtermInstance.current = term;

    term.writeln("Hey friend what's up.");

    let currentLine = '';
    const disposable = term.onData((data) => {
      const code = data.charCodeAt(0);

      if (code === 13) {
        term.write('\r\n');
        term.writeln(`You typed: ${currentLine}`);
        currentLine = '';
        term.write('$ ');
      } else if (code === 127) {
        if (currentLine.length > 0) {
          currentLine = currentLine.slice(0, -1);
          term.write('\b \b');
        }
      } else {
        currentLine += data;
        term.write(data);
      }
    });

    const handleResize = () => {
      fitAddon.fit();
    };
    window.addEventListener('resize', handleResize);

    // Cleanup logic when component unmounts
    return () => {
      disposable.dispose();
      window.removeEventListener('resize', handleResize);
      term.dispose();
    }
  });

  return (
    <div
      ref={terminalRef}
      style={{
        width: '100%',
        height: '400px',
        backgroundColor: '#1e1e1e',
        padding: '10px',
        boxSizing: 'border-box'
      }}
    />
  );
};
