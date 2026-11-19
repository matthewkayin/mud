import { type EquipmentSlot } from './socket_message';
import { MAX_ITEM_NAME_WIDTH, centerText, formatRow, formatViewHeader, truncate } from './view_format';

const EMPTY_SLOT_TEXT = '---';

// Builds the view panel's equipment table, one string per line
// The Item column is only as wide as its longest entry, and its cells are centered
export const formatEquipmentTable = (equipment: EquipmentSlot[]): string[] => {
  const itemCells = equipment.map((slot) => (slot.name === '' ? EMPTY_SLOT_TEXT : truncate(slot.name, MAX_ITEM_NAME_WIDTH)));

  const slotWidth = Math.max('Slot'.length, ...equipment.map((slot) => slot.slot.length));
  const itemWidth = Math.max('Item'.length, ...itemCells.map((cell) => cell.length));
  const widths = [slotWidth, itemWidth];
  // Each cell has a space of padding on both sides, and the columns are separated by a '|'
  const innerWidth = widths.reduce((total, width) => total + width + 2, 0) + widths.length - 1;

  const topBorder = formatViewHeader('Equipment', innerWidth + 2);
  const separator = `|${widths.map((width) => '-'.repeat(width + 2)).join('|')}|`;
  const bottomBorder = `+${'-'.repeat(innerWidth)}+`;

  const lines = [topBorder, formatRow(['Slot', centerText('Item', itemWidth)], widths), separator];
  equipment.forEach((slot, index) => {
    lines.push(formatRow([slot.slot, centerText(itemCells[index], itemWidth)], widths));
  });
  lines.push(bottomBorder);

  return lines;
};
