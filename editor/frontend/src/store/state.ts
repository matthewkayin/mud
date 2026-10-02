import { world } from '../../wailsjs/go/models';

export const GRID_INDEX_NONE = -1;
export const ROOM_NONE = -1;
export const DIRECTION_COUNT = 4;
export const WORLD_SECONDS_PER_UPDATE = 3;
export const CHEST_TYPE_CHEST = 0;

export type EditorGridCell = {
  x: number;
  y: number;
}

export type EditorGrid = {
  cellToRoomIndex: Map<string, number>;
  selectedCellKey: string | null;
}

export type EditorState = {
  rooms: world.Room[];
  grid: EditorGrid;
  itemData: world.ItemData[];
}

export function editorGridCellToString(cell: EditorGridCell): string {
  return `${cell.x},${cell.y}`;
}

export function editorGridCellFromString(key: string): EditorGridCell {
  const parts = key.split(',');
  return {
    x: Number.parseInt(parts[0]),
    y: Number.parseInt(parts[1]),
  }
}
