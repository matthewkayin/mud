import { type RoomMob } from './socket_message';
import { VIEW_WIDTH, formatRow, formatViewHeader, truncate } from './view_format';

// Shown when a mob isn't targeting a mob in the room
const EMPTY_CELL_TEXT = '---';

// Wide enough for the longest condition, 'Badly Wounded'
const CONDITION_WIDTH = 'Badly Wounded'.length;

// Builds the view panel's mobs table, one string per line, VIEW_WIDTH columns wide
// The Condition column fits the longest condition, and the Mob and Target columns split the rest, with longer names cut off
export const formatMobTable = (mobs: RoomMob[]): string[] => {
  // Each cell has a space of padding on both sides, and the table has a '|' on each side and between the columns
  const nameWidths = VIEW_WIDTH - CONDITION_WIDTH - 10;
  const mobWidth = Math.floor(nameWidths / 2);
  const targetWidth = nameWidths - mobWidth;
  const widths = [mobWidth, CONDITION_WIDTH, targetWidth];

  const topBorder = formatViewHeader('Mobs', VIEW_WIDTH);
  const separator = `|${widths.map((width) => '-'.repeat(width + 2)).join('|')}|`;
  const bottomBorder = `+${'-'.repeat(VIEW_WIDTH - 2)}+`;

  const lines = [topBorder, formatRow(['Mob', 'Condition', 'Target'], widths), separator];
  for (const mob of mobs) {
    const targetCell = mob.target === '' ? EMPTY_CELL_TEXT : truncate(mob.target, targetWidth);
    lines.push(formatRow([truncate(mob.name, mobWidth), mob.condition, targetCell], widths));
  }
  lines.push(bottomBorder);

  return lines;
};
