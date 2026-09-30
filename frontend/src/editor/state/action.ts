import * as mud from '../../mud/types';

export const EditorActionType = {
  ADD_ROOM: 0,
  EDIT_ROOM: 1,
} as const;
export type EditorActionType = (typeof EditorActionType)[keyof typeof EditorActionType];

export type EditorActionAddRoom = {
  roomGridIndex: number;
}

export type EditorActionEditRoom = {
  roomIndex: number;
  value: mud.Room;
  previous: mud.Room;
}

export type EditorAction = {
  type: EditorActionType,
  data: EditorActionAddRoom | EditorActionEditRoom,
}
