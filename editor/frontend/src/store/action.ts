import { world } from '../../wailsjs/go/models';
import { DIRECTION_COUNT, type EditorGridCell, editorGridCellToString, EditorState, ROOM_NONE } from './state';

export interface EditorAction {
  do: (state: EditorState) => void;
  undo: (state: EditorState) => void;
}

export class EditorActionAddRoom implements EditorAction {
  private data: { cellKey: string };

  constructor(cell: EditorGridCell) {
    this.data = {
      cellKey: editorGridCellToString(cell),
    };
  }

  do = (state: EditorState) => {
    const newRoomIndex = state.rooms.length;
    state.rooms.push(world.Room.createFrom({
      Name: 'New Room',
      Description: 'This is a new room',
      Exits: new Array(DIRECTION_COUNT).fill(ROOM_NONE),
      ExitIsLocked: new Array(DIRECTION_COUNT).fill(false),
      IsSafeZone: false,
      DropTable: {
        Entries: [],
      },
      Inventory: {
        Items: [],
      },
      Chests: [],
    }))

    state.grid.cellToRoomIndex.set(this.data.cellKey, newRoomIndex);
  }

  undo = (state: EditorState) => {
    state.rooms.pop();
    state.grid.cellToRoomIndex.delete(this.data.cellKey);
  }
}

export class EditorActionEditRoom implements EditorAction {
  private data: {
    roomIndex: number;
    previous: world.Room;
    value: world.Room;
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    state.rooms[this.data.roomIndex] = structuredClone(this.data.value);
  }

  undo = (state: EditorState) => {
    state.rooms[this.data.roomIndex] = structuredClone(this.data.previous);
  }
}
