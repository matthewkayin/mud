import { useState, useRef } from 'react';
import './terminal.css';

type TerminalProps = {
  prompt: string;
  lines: string[];
  command: string;
  setCommand: (value: string) => void;
  onSubmit: (command: string) => void;
}

export const Terminal = (props: TerminalProps) => {
  const [isFocused, setIsFocused] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

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
      if (props.command.length > 0) {
        props.onSubmit(props.command);
      }
      props.setCommand('');
      return;
    }

    if (event.key === 'Backspace') {
      props.setCommand(props.command.slice(0, -1))
      return;
    }

    if (event.key.length === 1 && event.key[0] >= ' ' && event.key[0] <= '~') {
      props.setCommand(props.command + event.key[0]);
    }
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
          {props.prompt} {props.command}
          <span className={isFocused ? "terminal-prompt-cursor-focused" : "terminal-prompt-cursor-hidden"}>&#x2588;</span></p>
      </div>
      <div className="terminal-contents-container">
        {props.lines.map((line: string, index) => (<p className="terminal-text" key={index}>{line}</p>))}
      </div>
    </div>
  )
};
