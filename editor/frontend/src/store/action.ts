import { world } from '../../wailsjs/go/models';
import { DIRECTION_COUNT, type EditorCell, EditorState, ROOM_NONE } from './state';

export interface EditorAction {
  do: (state: EditorState) => void;
  undo: (state: EditorState) => void;
}

export class EditorActionAddRoom implements EditorAction {
  private data: { cell: EditorCell };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    const key = this.data.cell.toString();
    state.rooms.set(key, world.Room.createFrom({
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
    }));
  }

  undo = (state: EditorState) => {
    const key = this.data.cell.toString();
    state.rooms.delete(key);
  }
}

export class EditorActionDeleteRoom implements EditorAction {
  private data: {
    cell: EditorCell;
    room: world.Room;
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    const key = this.data.cell.toString();
    state.rooms.delete(key);
  }

  undo = (state: EditorState) => {
    const key = this.data.cell.toString();
    state.rooms.set(key, structuredClone(this.data.room));
  }
}

export class EditorActionEditRoom implements EditorAction {
  private data: {
    cell: EditorCell;
    previous: world.Room;
    value: world.Room;
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    const key = this.data.cell.toString();
    state.rooms.set(key, structuredClone(this.data.value));
  }

  undo = (state: EditorState) => {
    const key = this.data.cell.toString();
    state.rooms.set(key, structuredClone(this.data.previous));
  }
}
