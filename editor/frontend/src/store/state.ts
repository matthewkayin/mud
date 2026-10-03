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

  isLessThanOrEqualTo = (other: EditorCell): boolean => {
    return this.x <= other.x && this.y <= other.y;
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
  connections: Map<string, string[]>;
  selectedCell: EditorCell | null;
  itemData: world.ItemData[];
}

export function editorStateDeleteRoom(state: EditorState, cell: EditorCell) {
  const key = cell.toString();

  // Remove key from any rooms that are connected to it
  const connections = state.connections.get(key) ?? [];
  for (const connKey of connections) {
    editorStateRemoveConnectionIfExists(state, connKey, key);
  }

  state.rooms.delete(key);
  state.connections.delete(key);
}

export function editorStateRemoveConnectionIfExists(state: EditorState, from: string, to: string) {
  const connections = state.connections.get(from);
  if (!connections) {
    return;
  }

  const index = connections.indexOf(to);
  if (index === -1) {
    return;
  }

  connections[index] = connections[connections.length - 1];
  connections.pop();
}
