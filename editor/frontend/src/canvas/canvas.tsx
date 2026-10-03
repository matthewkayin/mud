import { useRef, useEffect, useSyncExternalStore } from 'react';
import { Box, useTheme } from '@mui/material';
import { editorStore, EditorCell } from '../store';
import { world } from '../../wailsjs/go/models';
import { EditorActionAddRoom, EditorActionConnectRooms, EditorActionDisconnectRooms } from '../store/action';

const CAMERA_ZOOM_MIN = 0.25;
const CAMERA_ZOOM_MAX = 2.0;

const ROOM_X_SPACING = 30;
const ROOM_Y_SPACING = 30;
const ROOM_WIDTH = 200;
const ROOM_HEIGHT = 80;

type Rect = {
  x: number;
  y: number;
  width: number;
  height: number;
}

type CanvasState = {
  cameraOffsetX: number;
  cameraOffsetY: number;
  cameraZoom: number;

  mouseWorldX: number;
  mouseWorldY: number;
}

type CanvasColors = {
  background: string;
  foreground: string;
}

type RenderRoomParams = {
  x: number;
  y: number;
  dashBorder: boolean;
  selected: boolean;
  text: string;
}

export function Canvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const canvasBoxRef = useRef(null);
  const stateRef = useRef<CanvasState>({
    cameraOffsetX: 0,
    cameraOffsetY: 0,
    cameraZoom: 1.0,

    mouseWorldX: 0,
    mouseWorldY: 0,
  });

  // Theme colors
  const theme = useTheme();
  const colorsRef = useRef<CanvasColors>({
    background: theme.palette.background.default,
    foreground: theme.palette.text.primary,
  });
  useEffect(() => {
    colorsRef.current = {
      background: theme.palette.background.default,
      foreground: theme.palette.text.primary,
    };
  }, [theme]);

  // Editor selected cell slice
  const editorSelectedCell = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedCell());
  const editorSelectedCellRef = useRef(editorSelectedCell);
  useEffect(() => {
    editorSelectedCellRef.current = editorSelectedCell;
  }, [editorSelectedCell]);

  // Editor rooms slice
  const editorRooms = useSyncExternalStore(editorStore.subscribe, () => editorStore.getRooms());
  const editorRoomsRef = useRef(editorRooms);
  useEffect(() => {
    editorRoomsRef.current = editorRooms;
  }, [editorRooms]);

  // Editor connections slice
  const editorConnections = useSyncExternalStore(editorStore.subscribe, () => editorStore.getConnections());
  const editorConnectionsRef = useRef(editorConnections);
  useEffect(() => {
    editorConnectionsRef.current = editorConnections;
  }, [editorConnections]);

  // Event listeners
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) {
      return;
    }

    const onResize = () => {
      const rect = canvas.getBoundingClientRect();
      if (canvas.width !== rect.width || canvas.height !== rect.height) {
        canvas.width = rect.width;
        canvas.height = rect.height;
      }
    };
    onResize();

    // onContextMenu prevents the default right-click menu from popping up, allowing us
    // to use right-click to drag the camera
    const onContextMenu = (event: MouseEvent) => {
      event.preventDefault();
    };

    // MOUSE SCROLL
    const onMouseScroll = (event: WheelEvent) => {
      if (event.deltaY === 0) {
        return;
      }

      const state = stateRef.current;

      const canvasRect = canvas.getBoundingClientRect();
      const mouseX = event.clientX - canvasRect.left;
      const mouseY = event.clientY - canvasRect.top;

      const worldX = (mouseX - state.cameraOffsetX) / state.cameraZoom;
      const worldY = (mouseY - state.cameraOffsetY) / state.cameraZoom;

      if (event.deltaY < 0) {
        state.cameraZoom += state.cameraZoom * 0.1;
      } else if (event.deltaY > 0) {
        state.cameraZoom -= state.cameraZoom * 0.1;
      }

      if (state.cameraZoom < CAMERA_ZOOM_MIN) {
        state.cameraZoom = CAMERA_ZOOM_MIN;
      } else if (state.cameraZoom > CAMERA_ZOOM_MAX) {
        state.cameraZoom = CAMERA_ZOOM_MAX;
      }

      state.cameraOffsetX = mouseX - (worldX * state.cameraZoom);
      state.cameraOffsetY = mouseY - (worldY * state.cameraZoom);
    };

    // MOUSE CLICK
    const onMouseClick = () => {
      const rooms = editorRoomsRef.current;

      const hoveredCell = getHoveredCell(stateRef.current);
      if (hoveredCell) {
        const key = hoveredCell.toString();
        if (!rooms.has(key)) {
          editorStore.doAction(new EditorActionAddRoom({ cell: hoveredCell }));
        } else {
          editorStore.setSelectedCell(hoveredCell);
        }

        return;
      }

      const hoveredConnection = getHoveredConnection(stateRef.current, rooms);
      if (hoveredConnection) {
        const connections = editorConnectionsRef.current;

        if (!connectionExists(connections, hoveredConnection)) {
          editorStore.doAction(new EditorActionConnectRooms(hoveredConnection));
        } else {
          editorStore.doAction(new EditorActionDisconnectRooms(hoveredConnection));
        }
      }
    };

    // MOUSE MOVE
    const onMouseMove = (event: MouseEvent) => {
      const BUTTON_RIGHT = 2;
      const state = stateRef.current;

      // Mouse drag
      if ((event.buttons & BUTTON_RIGHT) === BUTTON_RIGHT) {
        state.cameraOffsetX += event.movementX;
        state.cameraOffsetY += event.movementY;
      }

      // Determine mouse world pos
      state.mouseWorldX = (event.offsetX - state.cameraOffsetX) / state.cameraZoom;
      state.mouseWorldY = (event.offsetY - state.cameraOffsetY) / state.cameraZoom;
    };

    window.addEventListener('resize', onResize);
    canvas.addEventListener('contextmenu', onContextMenu);
    canvas.addEventListener('click', onMouseClick);
    canvas.addEventListener('mousemove', onMouseMove);
    canvas.addEventListener('wheel', onMouseScroll);

    return () => {
      window.removeEventListener('resize', onResize);
      canvas.removeEventListener('contextmenu', onContextMenu);
      canvas.removeEventListener('click', onMouseClick);
      canvas.removeEventListener('mousemove', onMouseMove);
      canvas.removeEventListener('wheel', onMouseScroll);
    };
  }, []);

  // Draw
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) {
      return;
    }

    const context = canvas.getContext('2d');
    if (!context) {
      return;
    }

    let animationFrameId: number;
    const renderFrame = () => {
      render(context, stateRef.current, colorsRef.current, editorRoomsRef.current, editorConnectionsRef.current, editorSelectedCellRef.current);
      animationFrameId = requestAnimationFrame(renderFrame);
    };

    // start the render loop
    animationFrameId = requestAnimationFrame(renderFrame);

    return () => {
      cancelAnimationFrame(animationFrameId);
    };
  }, []);

  return (
    <Box
      ref={canvasBoxRef}
      sx={{
        display: 'block',
        width: '100%',
        height: '100%',
        backgroundColor: 'background.default',
      }}
    >
      <canvas
        ref={canvasRef}
        style={{
          width: 'auto',
          height: '100%',
          imageRendering: 'pixelated',
        }}
      />
    </Box>
  )
}

function render(context: CanvasRenderingContext2D, canvasState: CanvasState, colors: CanvasColors, rooms: Map<string, world.Room>, connections: Map<string, string[]>, selectedCell: EditorCell | null) {
  context.fillStyle = colors.background;
  context.fillRect(0, 0, context.canvas.width, context.canvas.height);

  context.save();
  context.setTransform(canvasState.cameraZoom, 0, 0, canvasState.cameraZoom, canvasState.cameraOffsetX, canvasState.cameraOffsetY);

  context.strokeStyle = colors.foreground;
  context.fillStyle = colors.foreground;

  // Render grid
  for (const [cellKey, room] of rooms) {
    const cell = EditorCell.fromString(cellKey);
    const isSelected = selectedCell ? selectedCell.isEqual(cell) : false;
    renderRoom(context, {
      x: cell.x,
      y: cell.y,
      dashBorder: false,
      selected: isSelected,
      text: room.Name,
    });
  }

  // Render connections
  for (const [cellKey, cellConnections] of connections) {
    // Get the cell
    const cell = EditorCell.fromString(cellKey);

    // Iterate through all connections
    for (const connKey of cellConnections) {
      const connCell = EditorCell.fromString(connKey);

      // Only render the connection if cell is less than connCell
      // This ensures that connections are only rendered once
      if (!cell.isLessThanOrEqualTo(connCell)) {
        continue;
      }

      renderConnection(context, cell, connCell, false);
    }
  }

  // New room hover
  const hoveredCell = getHoveredCell(canvasState);
  if (hoveredCell) {
    const hoveredCellKey = hoveredCell.toString();
    if (!rooms.has(hoveredCellKey)) {
      renderRoom(context, {
        x: hoveredCell.x,
        y: hoveredCell.y,
        dashBorder: true,
        selected: false,
        text: '+',
      });
    }
  }

  // Connection hover
  const hoveredConnection = getHoveredConnection(canvasState, rooms);
  if (hoveredConnection && !connectionExists(connections, hoveredConnection)) {
    renderConnection(context, hoveredConnection.from, hoveredConnection.to, true);
  }

  context.restore();
}

function renderRoom(context: CanvasRenderingContext2D, params: RenderRoomParams) {
  const rect = getRoomRect(params.x, params.y);

  if (params.dashBorder) {
    context.setLineDash([7, 2]);
  }
  context.strokeRect(rect.x + 0.5, rect.y + 0.5, rect.width, rect.height);

  if (params.selected) {
    context.setLineDash([7, 2]);
    context.strokeRect(rect.x - 5.5, rect.y - 5.5, rect.width + 11.5, rect.height + 11.5);
  }
  if (params.dashBorder || params.selected) {
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

function renderConnection(context: CanvasRenderingContext2D, fromRoom: EditorCell, toRoom: EditorCell, dashed: boolean) {
  let startX: number;
  let startY: number;
  let endX: number;
  let endY: number;

  const fromRect = getRoomRect(fromRoom.x, fromRoom.y);
  const toRect = getRoomRect(toRoom.x, toRoom.y);
  if (fromRoom.x === toRoom.x) {
    startX = fromRect.x + (fromRect.width / 2) - 0.5;
    endX = startX;

    if (fromRoom.y < toRoom.y) {
      startY = fromRect.y + fromRect.height;
      endY = toRect.y;
    } else {
      startY = toRect.y + toRect.height;
      endY = fromRect.y;
    }
  } else {
    startY = fromRect.y + (fromRect.height / 2) - 0.5;
    endY = startY;

    if (fromRoom.x < toRoom.x) {
      startX = fromRect.x + fromRect.width;
      endX = toRect.x;
    } else {
      startX = toRect.x + toRect.width;
      endX = fromRect.x;
    }
  }

  if (dashed) {
    context.setLineDash([7, 2]);
  }
  context.beginPath();
  context.moveTo(startX, startY);
  context.lineTo(endX, endY);
  context.stroke();
  if (dashed) {
    context.setLineDash([]);
  }
}

function getRoomRect(x: number, y: number): Rect {
  return {
    x: x * (ROOM_WIDTH + ROOM_X_SPACING),
    y: y * (ROOM_HEIGHT + ROOM_Y_SPACING),
    width: ROOM_WIDTH,
    height: ROOM_HEIGHT,
  };
}

function rectHasPoint(rect: Rect, x: number, y: number): boolean {
  return !(
    x < rect.x || y < rect.y || x >= rect.x + rect.width || y >= rect.y + rect.height
  );
}

function getHoveredCell(canvasState: CanvasState): EditorCell | null {
  const x = Math.floor(canvasState.mouseWorldX / (ROOM_WIDTH + ROOM_X_SPACING));
  const y = Math.floor(canvasState.mouseWorldY / (ROOM_HEIGHT + ROOM_Y_SPACING));
  const rect = getRoomRect(x, y);
  if (rectHasPoint(rect, canvasState.mouseWorldX, canvasState.mouseWorldY)) {
    return new EditorCell(x, y);
  }

  return null;
}

function getHoveredConnection(canvasState: CanvasState, rooms: Map<string, world.Room>): { from: EditorCell, to: EditorCell } | null {
  const x = Math.floor(canvasState.mouseWorldX / (ROOM_WIDTH + ROOM_X_SPACING));
  const y = Math.floor(canvasState.mouseWorldY / (ROOM_HEIGHT + ROOM_Y_SPACING));
  const cell = new EditorCell(x, y);

  if (!rooms.has(cell.toString())) {
    return null;
  }

  const rect = getRoomRect(x, y);

  // North connection
  if (canvasState.mouseWorldY < rect.y) {
    const adjacent = new EditorCell(x, y - 1);
    if (!rooms.has(adjacent.toString())) {
      return null;
    }

    return { from: cell, to: adjacent };
  }

  // East connection
  if (canvasState.mouseWorldX >= rect.x + rect.width) {
    const adjacent = new EditorCell(x + 1, y);
    if (!rooms.has(adjacent.toString())) {
      return null;
    }

    return { from: cell, to: adjacent };
  }

  // South connection
  if (canvasState.mouseWorldY >= rect.y + rect.height) {
    const adjacent = new EditorCell(x, y + 1);
    if (!rooms.has(adjacent.toString())) {
      return null;
    }

    return { from: cell, to: adjacent };
  }

  // West connection
  if (canvasState.mouseWorldX < rect.x) {
    const adjacent = new EditorCell(x - 1, y);
    if (!rooms.has(adjacent.toString())) {
      return null;
    }

    return { from: cell, to: adjacent };
  }

  return null;
}

function connectionExists(connections: Map<string, string[]>, hoveredConnection: { from: EditorCell, to: EditorCell }) {
  const fromKey = hoveredConnection.from.toString();
  const toKey = hoveredConnection.to.toString();
  return connections.get(fromKey)?.includes(toKey);
}
