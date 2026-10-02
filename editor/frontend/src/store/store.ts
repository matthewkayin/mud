import { EditorState } from './state';

export class EditorStore {
  private listeners = new Set<() => void>();
  private state: EditorState;

  constructor() {
    this.state = {
      rooms: [],
      grid: {
        width: 8,
        height: 8,
        roomIndices: new Array(8 * 8).fill(-1),
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

  getRooms = () => {
    return this.state.rooms;
  }

  getGrid = () => {
    return this.state.grid;
  }
}

export const editorStore = new EditorStore();
