import { type CharacterSheet, type SheetValue } from './socket_message';
import { VIEW_WIDTH, formatViewHeader } from './view_format';

// Each line's label is padded to this width, so the values line up
const LABEL_WIDTH = 'Experience'.length;

// The indent for lines that continue a labelled line, the width of a label and its ' : '
const CONTINUATION_INDENT = ' '.repeat(LABEL_WIDTH + 3);

const STAT_COLUMN_COUNT = 2;
const STAT_COLUMN_GAP = '   ';

const formatLabel = (label: string, value: string): string => `${label.padEnd(LABEL_WIDTH)} : ${value}`;

const formatSheetValue = (sheetValue: SheetValue): string => `${sheetValue.name} ${sheetValue.value}`;

// The stats in STAT_COLUMN_COUNT columns, each padded to its widest entry so the columns line up
const formatStatLines = (stats: SheetValue[]): string[] => {
  const entries = stats.map(formatSheetValue);
  const columnWidths = Array<number>(STAT_COLUMN_COUNT).fill(0);
  entries.forEach((entry, index) => {
    const column = index % STAT_COLUMN_COUNT;
    columnWidths[column] = Math.max(columnWidths[column], entry.length);
  });

  const lines: string[] = [];
  for (let index = 0; index < entries.length; index += STAT_COLUMN_COUNT) {
    const rowEntries = entries.slice(index, index + STAT_COLUMN_COUNT);
    const row = rowEntries.map((entry, column) => column === rowEntries.length - 1 ? entry : entry.padEnd(columnWidths[column])).join(STAT_COLUMN_GAP);
    lines.push(index === 0 ? formatLabel('Stats', row) : CONTINUATION_INDENT + row);
  }

  return lines;
};

// The skills as a comma-separated list, wrapped onto more lines so none is wider than VIEW_WIDTH
const formatSkillLines = (skills: string[]): string[] => {
  if (skills.length === 0) {
    return [formatLabel('Skills', 'None')];
  }

  const valueWidth = VIEW_WIDTH - CONTINUATION_INDENT.length;
  const rows: string[] = [];
  let row = '';
  skills.forEach((skill, index) => {
    const entry = index === skills.length - 1 ? skill : `${skill},`;
    if (row !== '' && row.length + 1 + entry.length > valueWidth) {
      rows.push(row);
      row = entry;
    } else {
      row = row === '' ? entry : `${row} ${entry}`;
    }
  });
  rows.push(row);

  return rows.map((rowText, index) => index === 0 ? formatLabel('Skills', rowText) : CONTINUATION_INDENT + rowText);
};

// Builds the view panel's stats mode, one string per line, with a VIEW_WIDTH header
export const formatCharacterSheet = (sheet: CharacterSheet): string[] => [
  formatViewHeader('Stats', VIEW_WIDTH),
  formatLabel('Name', `<y>${sheet.name}</y>`),
  formatLabel('Race', sheet.race),
  formatLabel('Class', sheet.class),
  '\n',
  formatLabel('Level', `${sheet.level}`),
  formatLabel('Experience', `${sheet.exp} / ${sheet.expNext}`),
  '\n',
  formatLabel('Attributes', sheet.attributes.map(formatSheetValue).join(' | ')),
  '\n',
  ...formatStatLines(sheet.stats),
  '\n',
  ...formatSkillLines(sheet.skills),
];
