import { world } from '../../wailsjs/go/models';

export const GRID_INDEX_NONE = -1;
export const ROOM_NONE = -1;
export const DIRECTION_COUNT = 4;
export const WORLD_SECONDS_PER_UPDATE = 3;
export const CHEST_TYPE_CHEST = 0;
export const WORLD_UPDATES_PER_MINUTE = 60 / WORLD_SECONDS_PER_UPDATE;

// These mirror the NpcDisposition and NpcMovementType enums in backend/world/npc.go
export const NPC_DISPOSITION_NAMES = ['Neutral', 'Hostile'];
export const NPC_DISPOSITION_NEUTRAL = 0;
export const NPC_MOVEMENT_TYPE_NAMES = ['Sentinel', 'Wander'];
export const NPC_MOVEMENT_TYPE_SENTINEL = 0;

export function minutesToTicks(minutes: number): number {
  return minutes * WORLD_UPDATES_PER_MINUTE;
}

export function ticksToMinutes(ticks: number): number {
  return ticks / WORLD_UPDATES_PER_MINUTE;
}

// The world stores NPCs in a world-level list that references rooms by index,
// but the editor stores them in the room they spawn in. NPC SpawnRoom values
// are filled in when the world is exported.
export type EditorRoom = world.Room & {
  Npcs: world.Npc[];
};

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

export class EditorConnection {
  from: EditorCell;
  to: EditorCell;

  constructor(paramFrom: EditorCell, paramTo: EditorCell) {
    this.from = paramFrom;
    this.to = paramTo;
  }

  isEqualTo = (other: EditorConnection): boolean => {
    return (this.from.isEqual(other.from) || this.from.isEqual(other.to)) &&
      (this.to.isEqual(other.from) || this.to.isEqual(other.to));
  }

  getDirection = (): number => {
    if (this.from.x === this.to.x && this.from.y - 1 === this.to.y) {
      return 0;
    }
    if (this.from.x + 1 === this.to.x && this.from.y === this.to.y) {
      return 1;
    }
    if (this.from.x === this.to.x && this.from.y + 1 === this.to.y) {
      return 2;
    }
    if (this.from.x - 1 === this.to.x && this.from.y === this.to.y) {
      return 3;
    }

    throw new Error("connection cells are not adjacent");
  }
}

export type EditorState = {
  rooms: Map<string, EditorRoom>;
  connections: Map<string, string[]>;
  selectedCell: EditorCell | null;
  selectedConnection: EditorConnection | null;
  itemData: world.ItemData[];
  npcData: world.NpcData[];
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

  if (state.selectedCell?.isEqual(cell)) {
    state.selectedCell = null;
  }
}

export function editorStateRemoveConnection(state: EditorState, connection: EditorConnection) {
  const fromKey = connection.from.toString();
  const toKey = connection.to.toString();

  editorStateRemoveConnectionIfExists(state, fromKey, toKey);
  editorStateRemoveConnectionIfExists(state, toKey, fromKey);

  if (state.selectedConnection?.isEqualTo(connection)) {
    state.selectedConnection = null;
  }
}

function editorStateRemoveConnectionIfExists(state: EditorState, from: string, to: string) {
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
