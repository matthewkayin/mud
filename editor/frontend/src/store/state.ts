import { world } from '../../wailsjs/go/models';

export const GRID_INDEX_NONE = -1;
export const ROOM_NONE = -1;
export const DIRECTION_COUNT = 4;
export const WORLD_SECONDS_PER_UPDATE = 3;
export const CHEST_TYPE_CHEST = 0;

export class EditorCell {
  x: number;
  y: number;

  constructor(paramX: number, paramY: number) {
    this.x = paramX;
    this.y = paramY;
  }

  isEqual = (other: EditorCell): boolean => {
    return this.x === other.x && this.y === other.y;
  }

  toString = (): string => {
    return `${this.x},${this.y}`;
  }

  static fromString = (cellString: string): EditorCell => {
    const parts = cellString.split(',');
    return new EditorCell(Number.parseInt(parts[0]), Number.parseInt(parts[1]));
  }
}

export type EditorState = {
  rooms: Map<string, world.Room>;
  selectedCell: EditorCell | null;
  itemData: world.ItemData[];
}
