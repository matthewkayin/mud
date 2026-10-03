import { world } from '../../wailsjs/go/models';
import {
  DIRECTION_COUNT,
  type EditorCell,
  EditorConnection,
  EditorState,
  editorStateDeleteRoom,
  editorStateRemoveConnection,
  ROOM_NONE
} from './state';

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
    connection: EditorConnection;
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    const fromKey = this.data.connection.from.toString();
    const toKey = this.data.connection.to.toString();

    state.connections.get(fromKey)?.push(toKey);
    state.connections.get(toKey)?.push(fromKey);
  }

  undo = (state: EditorState) => {
    editorStateRemoveConnection(state, this.data.connection);
  }
}

export class EditorActionDisconnectRooms implements EditorAction {
  private data: {
    connection: EditorConnection;
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    editorStateRemoveConnection(state, this.data.connection);
  }

  undo = (state: EditorState) => {
    const fromKey = this.data.connection.from.toString();
    const toKey = this.data.connection.to.toString();

    state.connections.get(fromKey)?.push(toKey);
    state.connections.get(toKey)?.push(fromKey);
  }
}

export class EditorActionEditConnection implements EditorAction {
  private data: {
    connection: EditorConnection;
    isLocked: boolean;
  };

  constructor(params: typeof this.data) {
    this.data = params;
  }

  do = (state: EditorState) => {
    this.setConnectionIsLocked(state, this.data.isLocked);
  }

  undo = (state: EditorState) => {
    this.setConnectionIsLocked(state, !this.data.isLocked);
  }

  private setConnectionIsLocked(state: EditorState, value: boolean) {
    const fromRoom = state.rooms.get(this.data.connection.from.toString());
    const toRoom = state.rooms.get(this.data.connection.to.toString());
    const direction = this.data.connection.getDirection();
    const reverseDirection = (direction + 2) % DIRECTION_COUNT;

    fromRoom!.ExitIsLocked[direction] = value;
    toRoom!.ExitIsLocked[reverseDirection] = value;
  }
}
