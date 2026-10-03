import { world } from '../../wailsjs/go/models';
import { DIRECTION_COUNT, type EditorCell, EditorState, editorStateDeleteRoom, editorStateRemoveConnectionIfExists, ROOM_NONE } from './state';

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

    state.connections.set(key, []);
  }

  undo = (state: EditorState) => {
    editorStateDeleteRoom(state, this.data.cell);
  }
}

export class EditorActionDeleteRoom implements EditorAction {
  private data: {
    cell: EditorCell;
    room: world.Room;
    connections: string[];
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    editorStateDeleteRoom(state, this.data.cell);
  }

  undo = (state: EditorState) => {
    const key = this.data.cell.toString();
    state.rooms.set(key, structuredClone(this.data.room));
    state.connections.set(key, structuredClone(this.data.connections));

    // Reconnect all connected rooms to key
    for (const connKey of this.data.connections) {
      state.connections.get(connKey)?.push(key);
    }
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

export class EditorActionConnectRooms implements EditorAction {
  private data: {
    from: EditorCell,
    to: EditorCell
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    const fromKey = this.data.from.toString();
    const toKey = this.data.to.toString();

    state.connections.get(fromKey)?.push(toKey);
    state.connections.get(toKey)?.push(fromKey);
  }

  undo = (state: EditorState) => {
    const fromKey = this.data.from.toString();
    const toKey = this.data.to.toString();

    editorStateRemoveConnectionIfExists(state, fromKey, toKey);
    editorStateRemoveConnectionIfExists(state, toKey, fromKey);
  }
}
