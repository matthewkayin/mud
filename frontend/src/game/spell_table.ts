import { type Spells } from './socket_message';
import { VIEW_WIDTH, formatRow, formatViewHeader, truncate } from './view_format';

// The table has a '|' and a space of padding on each side of its one column
const SPELL_COLUMN_WIDTH = VIEW_WIDTH - 4;

// A one-column bordered table of spell names, VIEW_WIDTH columns wide, with its title in the top border
const formatSpellTable = (title: string, spellNames: string[]): string[] => {
  const widths = [SPELL_COLUMN_WIDTH];

  const lines = [formatViewHeader(title, VIEW_WIDTH)];
  if (spellNames.length === 0) {
    lines.push(formatRow(['(none)'], widths));
  }
  for (const spellName of spellNames) {
    lines.push(formatRow([truncate(spellName, SPELL_COLUMN_WIDTH)], widths));
  }
  lines.push(`+${'-'.repeat(VIEW_WIDTH - 2)}+`);

  return lines;
};

// Builds the view panel's spells mode, one string per line: the prepared spells, with the used spell slots in the title, above the known spells
export const formatSpellTables = (spells: Spells): string[] => {
  return [
    ...formatSpellTable(`Prepared Spells (${spells.prepared.length} / ${spells.slots})`, spells.prepared),
    ...formatSpellTable('Known Spells', spells.known),
  ];
};
