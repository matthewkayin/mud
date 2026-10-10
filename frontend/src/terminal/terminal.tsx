import { useState, useRef, useEffect } from 'react';
import './terminal.css';

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
      if (command.length > 0) {
        onSubmit(command);
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

const terminalLineToSpan = (line: string, index: number) => {
  const parts = [];
  while (line.length !== 0) {
    const bracketIndex = line.indexOf("<");

    if (bracketIndex === -1) {
      parts.push({
        msg: line,
        color: "w"
      });
      line = "";
      continue;
    }

    if (bracketIndex > 0) {
      parts.push({
        msg: line.substring(0, bracketIndex),
        color: "w"
      });
      line = line.substring(bracketIndex);
    }

    const closeIndex = line.indexOf(">");
    if (closeIndex === -1 || closeIndex === line.length - 1) {
      parts.push({
        msg: line,
        color: "w"
      });
      line = "";
      continue;
    }

    const color = line[1];
    const closingTag = `</${color}>`;
    const closingTagIndex = line.indexOf(closingTag)

    if (closingTagIndex === -1) {
      parts.push({
        msg: line.substring(closeIndex + 1),
        color: color
      });
      line = "";
      continue;
    }

    parts.push({
      msg: line.substring(closeIndex + 1, closingTagIndex),
      color: color
    });
    const remainingIndex = closingTagIndex + closingTag.length;
    if (remainingIndex >= line.length) {
      line = "";
    } else {
      line = line.substring(remainingIndex);
    }
  }

  return (
    <span key = {index}>
      {parts.map((part) => {
        let color = "#fff";
        switch (part.color) {
          case "r":
            color = "#ff0000";
            break
          default:
            break
        }
        return (<span className = "terminal-text" style = {{ color }}>{part.msg}</span>)
      })}
    </span>
  );
}
