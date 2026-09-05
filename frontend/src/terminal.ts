import { useEffect, useRef, useCallback } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

export const useTerminal = (prompt = '> ') => {
  const containerRef = useRef(null);
  const terminalRef = useRef(null);
  const inputBufferRef = useRef('');
  const onSubmitRef = useRef(null);

  const writeLine = useCallback((text: string) => {
    if (terminalRef.current) {
      terminalRef.current.write(`${text}\r\n${prompt}`);
    }
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

    term.writeln('Hey friend.');
    term.write(prompt);

    const dataListener = term.onData((data) => {
      if (data === '\r') {
        const command = inputBufferRef.current.trim();
        term.write('\r\n');

        if (onSubmitRef.current && command) {
          onSubmitRef.current(command);
        } else {
          term.writeln('Warning: On submit not handled.');
          term.write(prompt);
        }
        inputBufferRef.current = '';

        return;
      }

      // Backspace
      if (data === '\u007F') {
        term.write('\b \b'); // Move cursor back, overwrite with space, and then move cursor back again
        inputBufferRef.current = inputBufferRef.current.slice(0, -1);
        return;
      }

      term.write(data);
      inputBufferRef.current += data;
    });

    const onResize = () => {
      fitAddon.fit();
    }
    window.addEventListener('resize', onResize);

    return () => {
      window.removeEventListener('resize', onResize);
      dataListener.dispose();
      term.dispose();
    };
  }, []);

  return { containerRef, writeLine, setOnSubmit };
};
