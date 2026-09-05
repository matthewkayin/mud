import { useEffect, useRef, useCallback } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

export const useTerminal = (prompt = '> ') => {
  const containerRef = useRef(null);
  const terminalRef = useRef(null);
  const inputBufferRef = useRef('');
  const onSubmitRef = useRef(null);

  const drawBottomPrompt = () => {
    if (!terminalRef.current) {
      return;
    }

    // Move cursor to the bottom row, first column
    terminalRef.current.write(`\x1b[${terminalRef.current.rows};1H`);

    // Clear the entire bottom line
    terminalRef.current.write('\x1b[2K');

    // Write prompt and current user input
    terminalRef.current.write(`${prompt} ${inputBufferRef.current}`);
  };


  const writeLine = useCallback((text: string) => {
    if (!terminalRef.current) {
      return;
    }

    terminalRef.current.write(`\x1b[${terminalRef.current.rows - 1};1H${text}\r\n`);
    drawBottomPrompt();
  }, []);

  const setOnSubmit = useCallback((callback) => {
    onSubmitRef.current = callback;
  }, []);

  useEffect(() => {
    if (!containerRef.current) {
      return;
    }

    const term = new Terminal({
      cursorBlink: true,
      theme: {
        background: '#1e1e1e',
        foreground: '#ffffff'
      }
    });
    term.open(containerRef.current);
    terminalRef.current = term;

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    fitAddon.fit();

    // Set the scrolling region to exclude the bottom row
    const totalRows = term.rows;
    const promptRow = totalRows - 1;
    term.write(`\x1b[1;${promptRow}r`);

    const dataListener = term.onData((data) => {
      // Submit
      if (data === '\r') {
        const command = inputBufferRef.current.trim();
        inputBufferRef.current = '';
        term.write('\r\n');

        if (onSubmitRef.current && command) {
          onSubmitRef.current(command);
        } else {
          writeLine('Warning: On submit not handled.');
        }

        return;
      }

      // Backspace
      if (data === '\u007F') {
        term.write('\b \b'); // Move cursor back, overwrite with space, and then move cursor back again
        inputBufferRef.current = inputBufferRef.current.slice(0, -1);
        return;
      }

      // Visible characters
      if (data >= ' ' && data <= '~') {
        term.write(data);
        inputBufferRef.current += data;
      }
    });

    const onResize = () => {
      fitAddon.fit();
    }
    window.addEventListener('resize', onResize);

    drawBottomPrompt();

    return () => {
      window.removeEventListener('resize', onResize);
      dataListener.dispose();
      term.dispose();
    };
  }, []);

  return { containerRef, writeLine, setOnSubmit };
};
