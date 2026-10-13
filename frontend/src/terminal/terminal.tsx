import { useState, useRef, useEffect } from 'react';
import './terminal.css';
import { parseLine } from './parser';
import { TEXT_COLOR_VALUES } from './text_color';

type TerminalProps = {
  prompt: string;
  lines: string[];
  command: string;
  setCommand: (value: string) => void;
  onSubmit: (command: string) => void;
}


export const Terminal = ({ prompt, lines, command, setCommand, onSubmit }: TerminalProps) => {
  const [isFocused, setIsFocused] = useState(false);
  const [shouldScrollToBottom, setShouldScrollToBottom] = useState(true);
  const containerRef = useRef<HTMLDivElement>(null);
  const terminalBottomElementRef = useRef<HTMLDivElement>(null);

  // Scroll to bottom
  useEffect(() => {
    if (terminalBottomElementRef.current === null) {
      return;
    }

    // Instead of keeping state, just check the scroll height here?
    if (shouldScrollToBottom) {
      terminalBottomElementRef.current.scrollIntoView();
    }
  }, [lines, shouldScrollToBottom]);

  const onFocus = () => {
    setIsFocused(true);
  };
  const onBlur = (event: React.FocusEvent<HTMLDivElement>) => {
    if (containerRef.current?.contains(event.relatedTarget as Node)) {
      return;
    }
    setIsFocused(false);
  };

  const onKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    if (!isFocused) {
      return;
    }

    // Handle submit
    if (event.key === 'Enter') {
      const sanitizedCommand = sanitizeLine(command);
      if (sanitizedCommand.length > 0) {
        onSubmit(sanitizedCommand);
        setShouldScrollToBottom(true);
      }
      setCommand('');
      return;
    }

    if (event.key === 'Backspace') {
      setCommand(command.slice(0, -1))
      return;
    }

    if (event.key.length === 1 && event.key[0] >= ' ' && event.key[0] <= '~') {
      setCommand(command + event.key[0]);
    }
  };

  const onScroll = (event: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = event.currentTarget;
    const lowestPossibleScrollTop = scrollHeight - clientHeight;
    const isScrolledToBottom = scrollTop == lowestPossibleScrollTop;
    setShouldScrollToBottom(isScrolledToBottom);
  };

  return (
    <div
      className="terminal"
      tabIndex={0}
      ref={containerRef}
      onFocus={onFocus}
      onBlur={onBlur}
      onKeyDown={onKeyDown}
    >
      <div className="terminal-prompt-container">
        <p className="terminal-text">
          {prompt} {command}
          <span className={isFocused ? "terminal-prompt-cursor-focused" : "terminal-prompt-cursor-hidden"}>&#x2588;</span></p>
      </div>
      <div
        className="terminal-contents-container"
        id="terminal-contents"
        onScroll={onScroll}
      >
        {lines.map(terminalLineToSpan)}
        <div ref={terminalBottomElementRef}></div>
      </div>
    </div>
  )
};

const sanitizeLine = (line: string) => {
  return line.replace(/[^\x20-\x7E]|[<>[\]{}\\]/g, "");
}

const terminalLineToSpan = (line: string, index: number) => {
  const tokens = parseLine(line);

  return (
    <p key={index} className="terminal-text">
      {tokens.map((token) => (<span style={{ color: TEXT_COLOR_VALUES[token.color] }}>{token.text}</span>) )}
    </p>
  );
}
