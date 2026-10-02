import { EditorGridCell, editorGridCellToString, EditorState, GRID_INDEX_NONE, ROOM_NONE } from './state';
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
      rooms: [],
      grid: {
        cellToRoomIndex: new Map<string, number>(),
        selectedCellKey: null,
      }
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

  getRooms = () => {
    return this.state.rooms;
  }

  getGrid = () => {
    return this.state.grid;
  }

  getSelectedRoomIndex = (): number | undefined => {
    if (this.state.grid.selectedCellKey === null) {
      return undefined;
    }
    return this.state.grid.cellToRoomIndex.get(this.state.grid.selectedCellKey);
  }

  setSelectedGridCell = (cell: EditorGridCell) => {
    this.state.grid.selectedCellKey = editorGridCellToString(cell);
    this.emitChange();
  }

  getSelectedRoom = (): world.Room | undefined => {
    if (this.state.grid.selectedCellKey === null) {
      return undefined;
    }
    const roomIndex = this.state.grid.cellToRoomIndex.get(this.state.grid.selectedCellKey);
    if (roomIndex === undefined) {
      return;
    }
    return this.state.rooms[roomIndex];
  }
}

export const editorStore = new EditorStore();
