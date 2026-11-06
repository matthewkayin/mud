import { TextColor } from './text_color';

export type TextToken = {
  text: string;
  color: TextColor;
}

type TextTag = {
  tag: string;
  textBefore: string;
  textInside: string;
  textAfter: string;
};

function findNextTag(line: string): TextTag | null {
  for (let index = 0; index < line.length; index++) {
    // Check for open bracket
    if (line[index] === '<') {
      // Now that we found an open bracket, find the closing bracket
      let closeBracketIndex = index + 1;
      while (closeBracketIndex < line.length && line[closeBracketIndex] !== '>') {
        closeBracketIndex++;
      }

      // If no closing bracket found, then this isn't a tag, and there is no open tag in the string
      if (closeBracketIndex === line.length) {
        return null;
      }

      // If the closing bracket is at the last char in the string, then there will be no closing tag, so this isn't an open tag
      if (closeBracketIndex + 1 === line.length) {
        return null;
      }

      // Once we found the closing bracket, get the tag name and make sure it's a valid tag
      const tag = index + 1 === closeBracketIndex ? '' : line.substring(index + 1, closeBracketIndex);
      const isValidTag = Object.values(TextColor).some((value) => value === tag);
      if (!isValidTag) {
        return null;
      }

      // Find closing tag
      const closingTag = `</${tag}>`;
      let closingTagIndex = line.substring(closeBracketIndex).indexOf(closingTag);

      // If there is no closing tag, then this isn't considered an opening tag
      if (closingTagIndex === -1) {
        return null;
      }
      closingTagIndex += closeBracketIndex;

      return {
        tag: tag,
        textBefore: index === 0 ? '' : line.substring(0, index),
        textInside: closingTagIndex === closeBracketIndex + 1 ? '' : line.substring(closeBracketIndex + 1, closingTagIndex),
        textAfter: closingTagIndex + closingTag.length >= line.length ? '' : line.substring(closingTagIndex + closingTag.length),
      };
    }
  }

  return null;
}

export function parseLine(line: string): TextToken[] {
  const tokens: TextToken[] = [];

  while (line.length !== 0) {
    // Find next open tag
    const tag = findNextTag(line);

    // If no tag found, take the rest of the line as a token
    if (!tag) {
      tokens.push({
        text: line,
        color: TextColor.WHITE,
      });
      break;
    }

    // Take part before tag
    if (tag.textBefore.length !== 0) {
      tokens.push({
        text: tag.textBefore,
        color: TextColor.WHITE
      });
    }

    // Take part inside the tag
    if (tag.textInside.length !== 0) {
      tokens.push({
        text: tag.textInside,
        color: tag.tag as TextColor,
      });
    }

    // Set line equal to the part after the tag
    line = tag.textAfter;
  }

  return tokens;
}
