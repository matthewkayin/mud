import { EditorCell, EditorConnection, EditorRoom, EditorState } from './state';
import { EditorAction } from './action';
import { main, world } from '../api/models';
import { SetIsDirty } from '../api/editor_api';

const ACTION_HISTORY_MAX_LENGTH = 64;

export class EditorStore {
  private listeners = new Set<() => void>();
  private state: EditorState;

  private actionHistory: EditorAction[] = [];
  private actionHistoryIndex = 0;

  // The action at the top of the history when the world was last saved or opened.
  // Actions are compared by identity, so this still works after old history is trimmed.
  private savedAction: EditorAction | null = null;
  private isDirty = false;

  constructor() {
    this.state = {
      rooms: new Map<string, EditorRoom>(),
      connections: new Map<string, string[]>(),
      selectedCell: null,
      selectedConnection: null,
      startCell: null,
      itemData: [],
      npcData: [],
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
    while (this.actionHistory.length >= ACTION_HISTORY_MAX_LENGTH) {
      this.actionHistory.splice(0, 1);
      this.actionHistoryIndex--;
    }
    this.actionHistory.push(action);
    this.actionHistoryIndex++;
    this.updateIsDirty();
    this.emitChange();
  }

  undoAction = () => {
    if (this.actionHistoryIndex === 0) {
      return;
    }

    this.actionHistoryIndex--;
    const action = this.actionHistory[this.actionHistoryIndex];
    action.undo(this.state);
    this.updateIsDirty();
    this.emitChange();
  }

  redoAction = () => {
    if (this.actionHistoryIndex === this.actionHistory.length) {
      return;
    }

    const action = this.actionHistory[this.actionHistoryIndex];
    action.do(this.state);
    this.actionHistoryIndex++;
    this.updateIsDirty();
    this.emitChange();
  }

  getTopAction = (): EditorAction | null => {
    return this.actionHistoryIndex > 0 ? this.actionHistory[this.actionHistoryIndex - 1] : null;
  }

  private updateIsDirty = () => {
    const isDirty = this.getTopAction() !== this.savedAction;
    if (isDirty !== this.isDirty) {
      this.isDirty = isDirty;
      SetIsDirty(isDirty);
    }
  }

  markSaved = (savePoint: EditorAction | null) => {
    this.savedAction = savePoint;
    this.updateIsDirty();
  }

  // Converts the editor state into the format the Go side saves
  toEditorWorld = (): main.EditorWorld => {
    const rooms = [];
    for (const [key, editorRoom] of this.state.rooms) {
      const { Npcs, ...room } = editorRoom;
      const connections = this.state.connections.get(key) ?? [];
      rooms.push({
        Position: EditorCell.fromString(key).toPosition(),
        Room: room,
        Npcs: Npcs,
        Connections: connections.map((connKey) => EditorCell.fromString(connKey).toPosition()),
      });
    }

    // This is a plain object rather than main.EditorWorld.createFrom() because createFrom
    // would rebuild each NPC's Behavior from its generated model, which drops the behavior JSON.
    return {
      Rooms: rooms,
      StartRoom: this.state.startCell?.toPosition(),
    } as main.EditorWorld;
  }

  // Replaces the editor state with an opened world. This can't be undone.
  loadEditorWorld = (editorWorld: main.EditorWorld) => {
    const rooms = new Map<string, EditorRoom>();
    const connections = new Map<string, string[]>();
    for (const editorRoom of editorWorld.Rooms) {
      const key = EditorCell.fromPosition(editorRoom.Position).toString();
      rooms.set(key, Object.assign(editorRoom.Room, { Npcs: editorRoom.Npcs }));
      connections.set(key, editorRoom.Connections.map((position) => EditorCell.fromPosition(position).toString()));
    }

    this.state.rooms = rooms;
    this.state.connections = connections;
    this.state.startCell = editorWorld.StartRoom ? EditorCell.fromPosition(editorWorld.StartRoom) : null;
    this.state.selectedCell = null;
    this.state.selectedConnection = null;

    this.actionHistory = [];
    this.actionHistoryIndex = 0;
    this.savedAction = null;
    this.updateIsDirty();
    this.emitChange();
  }

  getItemData = () => {
    return this.state.itemData;
  }

  setItemData = (itemData: world.ItemData[]) => {
    this.state.itemData = itemData;
    this.emitChange();
  }

  getNpcData = () => {
    return this.state.npcData;
  }

  setNpcData = (npcData: world.NpcData[]) => {
    this.state.npcData = npcData;
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

  getSelectedRoom = (): EditorRoom | undefined => {
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

  getRooms = (): Map<string, EditorRoom> => {
    return this.state.rooms;
  }

  getConnections = (): Map<string, string[]> => {
    return this.state.connections;
  }

  getStartCell = (): EditorCell | null => {
    return this.state.startCell;
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
