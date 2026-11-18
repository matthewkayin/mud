import { parseLine } from './parser';
import { TEXT_COLOR_VALUES } from './text_color';

// Renders a line of color-tagged text as a terminal-styled paragraph
export const textLineToParagraph = (line: string, index: number) => {
  const tokens = parseLine(line);

  return (
    <p key={index} className="terminal-text">
      {tokens.map((token, tokenIndex) => (<span key={tokenIndex} style={{ color: TEXT_COLOR_VALUES[token.color] }}>{token.text}</span>))}
    </p>
  );
};
