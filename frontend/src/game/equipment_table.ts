import { type EquipmentSlot } from './socket_message';
import { MAX_ITEM_NAME_WIDTH, centerText, formatRow, formatViewHeader, truncate } from './view_format';

// Shown for an empty slot, and for the slot name of the off hand while a two-handed weapon is equipped
const EMPTY_CELL_TEXT = '---';

// Builds the view panel's equipment table, one string per line, tableWidth columns wide
// The Slot column fits the longest slot name, and the Item column stretches to fill the rest, with its cells centered
export const formatEquipmentTable = (equipment: EquipmentSlot[], tableWidth: number): string[] => {
  const slotCells = equipment.map((slot) => (slot.slot === '' ? EMPTY_CELL_TEXT : slot.slot));
  const itemCells = equipment.map((slot) => (slot.name === '' ? EMPTY_CELL_TEXT : truncate(slot.name, MAX_ITEM_NAME_WIDTH)));

  const slotWidth = Math.max('Slot'.length, ...slotCells.map((cell) => cell.length));
  // Each cell has a space of padding on both sides, and the table has a '|' on each side and between the columns
  const itemWidth = Math.max('Item'.length, tableWidth - slotWidth - 7);
  const widths = [slotWidth, itemWidth];
  const innerWidth = slotWidth + itemWidth + 5;

  const topBorder = formatViewHeader('Equipment', innerWidth + 2);
  const separator = `|${widths.map((width) => '-'.repeat(width + 2)).join('|')}|`;
  const bottomBorder = `+${'-'.repeat(innerWidth)}+`;

  const lines = [topBorder, formatRow(['Slot', centerText('Item', itemWidth)], widths), separator];
  equipment.forEach((slot, index) => {
    const slotCell = slot.slot === '' ? centerText(slotCells[index], slotWidth) : slotCells[index];
    lines.push(formatRow([slotCell, centerText(itemCells[index], itemWidth)], widths));
  });
  lines.push(bottomBorder);

  return lines;
};
