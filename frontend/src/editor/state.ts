import { type World } from '../mud/types';
import { type Vec2, vec2Subtract } from './vec2';

const RENDER_ROOM_WIDTH = 100;
const RENDER_ROOM_HEIGHT = 40;

export class EditorState {
  world: World = {
    Rooms: [],
    Npcs: [],
  };
  roomPositions: Vec2[] = [];
  newRoomPositions: Vec2[] = [
    { x: 10, y: 10 }
  ];
  cameraOffset: Vec2 = { x: 0, y: 0 };

  constructor() {
  }

  setCameraOffset() {
  }

  render(context: CanvasRenderingContext2D) {
    context.fillStyle = '#000';
    context.fillRect(0, 0, context.canvas.width, context.canvas.height);

    context.strokeStyle = 'white';
    this.newRoomPositions.forEach((position) => {
      this.renderRoom(context, vec2Subtract(position, this.cameraOffset), "+");
    });

  }

  renderRoom(context: CanvasRenderingContext2D, position: Vec2, text: string) {
    context.strokeRect(position.x + 0.5, position.y + 0.5, RENDER_ROOM_WIDTH, RENDER_ROOM_HEIGHT);
  }
}
