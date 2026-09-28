import * as mud from '../mud/types';
import { type Vec2, type Rect, rectHasPoint } from './util';

const ROOM_X_SPACING = 20;
const ROOM_Y_SPACING = 20;
const ROOM_WIDTH = 200;
const ROOM_HEIGHT = 80;

export const EDITOR_ROOM_GRID_INDEX_NONE = -1;

export const EditorStateHoverType = {
  NONE: 0,
  ROOM: 1,
  NEW_ROOM: 2,
} as const;
export type EditorStateHoverType = (typeof EditorStateHoverType)[keyof typeof EditorStateHoverType];

export type EditorStateHover = {
  type: EditorStateHoverType;
  index: number;
}

type RenderRoomParams = {
  cell: Vec2;
  dashBorder: boolean;
  color: string;
  text: string;
}

export class EditorState {
  cameraOffset: Vec2 = { x: 0, y: 0 };
  cameraZoom: number = 1.0;
  hoveredRoomGridIndex: number = EDITOR_ROOM_GRID_INDEX_NONE;
  selectedRoomGridIndex: number = EDITOR_ROOM_GRID_INDEX_NONE;

  world: mud.World = {
    Rooms: [],
    Npcs: [],
  };

  roomGridWidth: number;
  roomGridHeight: number;
  roomGrid: number[] = [];

  constructor() {
    this.roomGridWidth = 16;
    this.roomGridHeight = 16;
    this.roomGrid = new Array(this.roomGridWidth * this.roomGridHeight).fill(mud.ROOM_NONE);
  }

  roomGridIndex(cell: Vec2): number {
    return Math.floor(cell.x + (cell.y * this.roomGridWidth));
  }

  roomGridCellFromIndex(index: number): Vec2 {
    return {
      x: index % this.roomGridWidth,
      y: Math.floor(index / this.roomGridWidth),
    };
  }

  setRoomGrid(index: number, value: number) {
    const cell = this.roomGridCellFromIndex(index);
    this.roomGrid[index] = value;

    // If we just placed a room at the edge of the grid, resize the grid

    // Determine direction to resize
    const growthAmount = 1;
    let shiftX = 0;
    let shiftY = 0;
    if (cell.x === 0) {
      shiftX = growthAmount;
    }
    if (cell.x === this.roomGridWidth - 1) {
      shiftX = -growthAmount;
    }
    if (cell.y === 0) {
      shiftY = growthAmount;
    }
    if (cell.y === this.roomGridHeight - 1) {
      shiftY = -growthAmount;
    }

    // If no direction to resize, then don't resize
    if (shiftX === 0 && shiftY === 0) {
      return;
    }

    // Alloc new grid
    const newWidth = this.roomGridWidth + Math.abs(shiftX);
    const newHeight = this.roomGridHeight + Math.abs(shiftY);
    const newGrid = new Array(newWidth * newHeight).fill(mud.ROOM_NONE);

    // Copy old grid into new grid
    for (let y = 0; y < this.roomGridHeight; y++) {
      for (let x = 0; x < this.roomGridWidth; x++) {
        const index = this.roomGridIndex({ x, y });
        const newIndex = (x + shiftX) + ((y + shiftY) * newWidth);

        newGrid[newIndex] = this.roomGrid[index];
      }
    }

    // Save new grid to class
    this.roomGridWidth = newWidth;
    this.roomGridHeight = newHeight;
    this.roomGrid = newGrid;

    // Adjust camera
    this.cameraOffset.x -= shiftX * (ROOM_WIDTH + ROOM_X_SPACING);
    this.cameraOffset.y -= shiftY * (ROOM_HEIGHT + ROOM_Y_SPACING);
  }

  getRoomRect(cell: Vec2): Rect {
    return {
      x: cell.x * (ROOM_WIDTH + ROOM_X_SPACING),
      y: cell.y * (ROOM_HEIGHT + ROOM_Y_SPACING),
      width: ROOM_WIDTH,
      height: ROOM_HEIGHT,
    };
  }

  onMouseMoved(mouseWorldPos: Vec2) {
    this.hoveredRoomGridIndex = EDITOR_ROOM_GRID_INDEX_NONE;

    // Check for hovering over rooms
    for (let y = 0; y < this.roomGridHeight; y++) {
      for (let x = 0; x < this.roomGridWidth; x++) {
        const roomRect = this.getRoomRect({ x, y });
        if (rectHasPoint(roomRect, mouseWorldPos)) {
          this.hoveredRoomGridIndex = this.roomGridIndex({ x, y });
          return;
        }
      }
    }
  }

  createRoom() {
    if (this.hoveredRoomGridIndex === EDITOR_ROOM_GRID_INDEX_NONE) {
      return;
    }
    if (this.roomGrid[this.hoveredRoomGridIndex] !== mud.ROOM_NONE) {
      return;
    }

    const newRoomIndex = this.world.Rooms.length;
    this.world.Rooms.push({
      Name: 'New Room',
      Description: 'This is a new room',
      Exits: new Array(mud.DIRECTION_COUNT).fill(mud.ROOM_NONE),
      ExitIsLocked: new Array(mud.DIRECTION_COUNT).fill(false),
      IsSafeZone: false,
      Inventory: {
        Items: [],
      },
      Chests: [],
    });

    this.setRoomGrid(this.hoveredRoomGridIndex, newRoomIndex);
  }

  render(context: CanvasRenderingContext2D) {
    context.fillStyle = '#000';
    context.fillRect(0, 0, context.canvas.width, context.canvas.height);

    context.save();
    context.translate(this.cameraOffset.x, this.cameraOffset.y);
    context.scale(this.cameraZoom, this.cameraZoom);

    // Render rooms
    for (let y = 0; y < this.roomGridHeight; y++) {
      for (let x = 0; x < this.roomGridWidth; x++) {
        const roomGridIndex = this.roomGridIndex({ x, y });
        const roomIndex = this.roomGrid[roomGridIndex];
        if (roomIndex === mud.ROOM_NONE) {
          continue;
        }

        const room = this.world.Rooms[roomIndex];
        const isSelected = this.selectedRoomGridIndex === roomGridIndex;
        this.renderRoom(context, {
          cell: { x, y },
          dashBorder: false,
          color: isSelected ? '#ffff00' : '#fff',
          text: room.Name,
        });
      }
    }

    // Render new room hover
    const hoveredRoomIndex = this.roomGrid[this.hoveredRoomGridIndex];
    if (hoveredRoomIndex === mud.ROOM_NONE) {
      this.renderRoom(context, {
        cell: this.roomGridCellFromIndex(this.hoveredRoomGridIndex),
        dashBorder: true,
        color: '#fff',
        text: '+',
      });
    }

    context.restore();
  }

  renderRoom(context: CanvasRenderingContext2D, params: RenderRoomParams) {
    context.strokeStyle = params.color;
    context.fillStyle = params.color;

    const rect = this.getRoomRect(params.cell);

    if (params.dashBorder) {
      context.setLineDash([7, 2]);
    }
    context.strokeRect(rect.x + 0.5, rect.y + 0.5, rect.width, rect.height);
    if (params.dashBorder) {
      // Reset line dash
      context.setLineDash([])
    }

    context.font = '16px sans-serif';
    context.textAlign = 'center';
    context.textBaseline = 'middle';

    const centerX = rect.x + (rect.width / 2);
    const centerY = rect.y + (rect.height / 2);

    context.fillText(params.text, centerX, centerY);
  }
}
