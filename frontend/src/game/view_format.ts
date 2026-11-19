// The view panel's standard line width, which is also the inventory table's width for the default capacity
export const VIEW_WIDTH = 70;

// Wide enough for "Damaged Sword of the Forgotten Kingdom"; longer item names in the view panel's tables are truncated
export const MAX_ITEM_NAME_WIDTH = 38;

const TRUNCATION_SUFFIX = '...';

// A '+---- Title ----+' line, width columns wide, with the title centered
export const formatViewHeader = (title: string, width: number): string => {
  const paddedTitle = ` ${title} `;
  const dashCount = Math.max(width - 2 - paddedTitle.length, 0);
  const leftDashCount = Math.floor(dashCount / 2);
  const rightDashCount = dashCount - leftDashCount;
  return `+${'-'.repeat(leftDashCount)}${paddedTitle}${'-'.repeat(rightDashCount)}+`;
};

// Cuts text longer than width off with a trailing '...'
export const truncate = (text: string, width: number): string => {
  if (text.length <= width) {
    return text;
  }
  return text.slice(0, width - TRUNCATION_SUFFIX.length) + TRUNCATION_SUFFIX;
};

// Pads text with spaces on both sides to width columns, with any odd space on the right
export const centerText = (text: string, width: number): string => {
  const leftPadding = Math.floor(Math.max(width - text.length, 0) / 2);
  return (' '.repeat(leftPadding) + text).padEnd(width);
};

// A '| cell | cell |' table row, with each cell left-aligned in its column's width
export const formatRow = (cells: string[], widths: number[]): string => {
  const paddedCells = cells.map((cell, index) => ` ${cell.padEnd(widths[index])} `);
  return `|${paddedCells.join('|')}|`;
};
