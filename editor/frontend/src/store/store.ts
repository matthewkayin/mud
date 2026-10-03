import { EditorCell, EditorConnection, EditorState, GRID_INDEX_NONE, ROOM_NONE } from './state';
import { EditorAction } from './action';
import { world } from '../../wailsjs/go/models';

const ACTION_HISTORY_MAX_LENGTH = 64;

export class EditorStore {
  private listeners = new Set<() => void>();
  private state: EditorState;

  private actionHistory: EditorAction[] = [];
  private actionHistoryIndex = 0;

  constructor() {
    this.state = {
      rooms: new Map<string, world.Room>(),
      connections: new Map<string, string[]>(),
      selectedCell: null,
      selectedConnection: null,
      itemData: [],
    };
  }

  // Adds a new listener who will subscribe to updates
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  // Notify all listeners that a change has been made
  emitChange = () => {
    this.listeners.forEach((listener) => listener());
  }

  doAction = (action: EditorAction) => {
    action.do(this.state);

    while (this.actionHistory.length > this.actionHistoryIndex) {
      this.actionHistory.pop();
    }
    while (this.actionHistory.length > ACTION_HISTORY_MAX_LENGTH) {
      this.actionHistory.splice(0, 1);
    }
    this.actionHistory.push(action);
    this.actionHistoryIndex++;
    this.emitChange();
  }

  undoAction = () => {
    if (this.actionHistoryIndex === 0) {
      return;
    }

    this.actionHistoryIndex--;
    const action = this.actionHistory[this.actionHistoryIndex];
    action.undo(this.state);
    this.emitChange();
  }

  redoAction = () => {
    if (this.actionHistoryIndex === this.actionHistory.length) {
      return;
    }

    const action = this.actionHistory[this.actionHistoryIndex];
    action.do(this.state);
    this.actionHistoryIndex++;
    this.emitChange();
  }

  getItemData = () => {
    return this.state.itemData;
  }

  setItemData = (itemData: world.ItemData[]) => {
    this.state.itemData = itemData;
    this.emitChange();
  }

  getSelectedCell = (): EditorCell | null => {
    return this.state.selectedCell;
  }

  setSelectedCell = (cell: EditorCell) => {
    this.state.selectedCell = cell;
    this.state.selectedConnection = null;
    this.emitChange();
  }

  getSelectedConnection = (): EditorConnection | null => {
    return this.state.selectedConnection;
  }

  setSelectedConnection = (connection: EditorConnection) => {
    this.state.selectedConnection = connection;
    this.state.selectedCell = null;
    this.emitChange();
  }

  getSelectedRoom = (): world.Room | undefined => {
    if (this.state.selectedCell === null) {
      return undefined;
    }

    return this.state.rooms.get(this.state.selectedCell.toString());
  }

  getSelectedRoomConnections = (): string[] | undefined => {
    if (this.state.selectedCell === null) {
      return undefined;
    }

    return this.state.connections.get(this.state.selectedCell.toString());
  }

  getRooms = (): Map<string, world.Room> => {
    return this.state.rooms;
  }

  getConnections = (): Map<string, string[]> => {
    return this.state.connections;
  }

  getSelectedConnectionIsLocked = (): boolean => {
    if (!this.state.selectedConnection) {
      return false;
    }

    const fromRoom = this.state.rooms.get(this.state.selectedConnection.from.toString());
    const direction = this.state.selectedConnection.getDirection();
    return fromRoom!.ExitIsLocked[direction];
  }
}

export const editorStore = new EditorStore();
