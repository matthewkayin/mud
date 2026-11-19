import { type Inventory } from './socket_message';
import { MAX_ITEM_NAME_WIDTH, formatRow, formatViewHeader, truncate } from './view_format';

const AMOUNT_COLUMN_WIDTH = 'Amount'.length;
// Wide enough for "Size (100 / 100)", so the column doesn't change width as the inventory fills
const SIZE_COLUMN_MIN_WIDTH = 'Size (100 / 100)'.length;

// Builds the view panel's inventory table, one string per line
export const formatInventoryTable = (inventory: Inventory): string[] => {
  const sizeHeader = `Size (${inventory.size} / ${inventory.capacity})`;
  const widths = [MAX_ITEM_NAME_WIDTH, AMOUNT_COLUMN_WIDTH, Math.max(SIZE_COLUMN_MIN_WIDTH, sizeHeader.length)];
  // Each cell has a space of padding on both sides, and the columns are separated by a '|'
  const innerWidth = widths.reduce((total, width) => total + width + 2, 0) + widths.length - 1;

  const topBorder = formatViewHeader('Inventory', innerWidth + 2);
  const separator = `|${widths.map((width) => '-'.repeat(width + 2)).join('|')}|`;
  const bottomBorder = `+${'-'.repeat(innerWidth)}+`;

  const lines = [topBorder, formatRow(['Item', 'Amount', sizeHeader], widths), separator];
  if (inventory.items.length === 0) {
    lines.push(formatRow(['(empty)', '', ''], widths));
  }
  for (const item of inventory.items) {
    lines.push(formatRow([truncate(item.name, MAX_ITEM_NAME_WIDTH), `x${item.amount}`, `${item.size}`], widths));
  }
  lines.push(bottomBorder);

  return lines;
};
