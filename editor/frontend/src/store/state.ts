import { world } from '../../wailsjs/go/models';

export type EditorGrid = {
  width: number;
  height: number;
  roomIndices: number[];
}

export type EditorState = {
  rooms: world.Room[];
  grid: EditorGrid;
}
