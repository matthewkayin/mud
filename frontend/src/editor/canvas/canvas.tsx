import { useRef, useEffect } from 'react';
import { Box, useTheme } from '@mui/material';
import { editorStore, EditorCell, EditorConnection, type EditorRoom } from '../store';
import { EditorActionAddRoom, EditorActionConnectRooms } from '../store/action';

const CAMERA_ZOOM_MIN = 0.25;
const CAMERA_ZOOM_MAX = 2.0;

const ROOM_X_SPACING = 30;
const ROOM_Y_SPACING = 30;
const ROOM_WIDTH = 200;
const ROOM_HEIGHT = 80;

const RenderStyle = {
  SOLID: 0,
  DASHED: 1,
  SELECTED: 2,
} as const;
type RenderStyle = typeof RenderStyle[keyof typeof RenderStyle];

type Rect = {
  x: number;
  y: number;
  width: number;
  height: number;
}

type CellRange = {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
}

type CanvasState = {
  cameraOffsetX: number;
  cameraOffsetY: number;
  cameraZoom: number;

  isMouseOver: boolean;
  mouseWorldX: number;
  mouseWorldY: number;

  // Identifies what the mouse is hovering, so mouse moves only redraw when it changes
  hoverKey: string;
}

type CanvasColors = {
  background: string;
  foreground: string;
}

type RenderRoomParams = {
  x: number;
  y: number;
  renderStyle: RenderStyle;
  text: string;
  isStartRoom: boolean;
}

const ROOM_FONT = '16px sans-serif';
const START_LABEL_FONT = '12px sans-serif';

export function Canvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const canvasBoxRef = useRef<HTMLDivElement>(null);
  const stateRef = useRef<CanvasState>({
    cameraOffsetX: 0,
    cameraOffsetY: 0,
    cameraZoom: 1.0,

    isMouseOver: false,
    mouseWorldX: 0,
    mouseWorldY: 0,

    hoverKey: '',
  });

  // The canvas only redraws when something changes, so anything that affects
  // the picture must call this
  const requestRenderRef = useRef<(() => void) | null>(null);

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
    requestRenderRef.current?.();
  }, [theme]);

  useEffect(() => {
    const canvas = canvasRef.current;
    const canvasBox = canvasBoxRef.current;
    if (!canvas || !canvasBox) {
      return;
    }

    const context = canvas.getContext('2d');
    if (!context) {
      return;
    }

    // Draw
    let animationFrameId = 0;
    let isFrameRequested = false;
    const requestRender = () => {
      if (isFrameRequested) {
        return;
      }

      isFrameRequested = true;
      animationFrameId = requestAnimationFrame(() => {
        isFrameRequested = false;
        render(context, stateRef.current, colorsRef.current);
      });
    };
    requestRenderRef.current = requestRender;

    // Redraw whenever the editor state changes
    const unsubscribe = editorStore.subscribe(requestRender);

    // The hover preview depends on the rooms, so the hover key is refreshed on every
    // mouse event and a redraw is only needed when it changes
    const updateHover = () => {
      const state = stateRef.current;
      const hoverKey = getHoverKey(state, editorStore.getRooms());
      if (hoverKey !== state.hoverKey) {
        state.hoverKey = hoverKey;
        requestRender();
      }
    };

    // Resizing the canvas clears it, so it always needs a redraw
    const onResize = () => {
      const width = canvasBox.clientWidth;
      const height = canvasBox.clientHeight;
      if (canvas.width !== width || canvas.height !== height) {
        canvas.width = width;
        canvas.height = height;
      }
      requestRender();
    };
    const resizeObserver = new ResizeObserver(onResize);
    resizeObserver.observe(canvasBox);
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

      requestRender();
    };

    // MOUSE CLICK
    const onMouseClick = () => {
      const rooms = editorStore.getRooms();

      const hoveredCell = getHoveredCell(stateRef.current);
      if (hoveredCell) {
        const key = hoveredCell.toString();
        if (!rooms.has(key)) {
          editorStore.doAction(new EditorActionAddRoom({ cell: hoveredCell }));
        }
        editorStore.setSelectedCell(hoveredCell);

        return;
      }

      const hoveredConnection = getHoveredConnection(stateRef.current, rooms);
      if (hoveredConnection) {
        const connections = editorStore.getConnections();

        if (!connectionExists(connections, hoveredConnection)) {
          editorStore.doAction(new EditorActionConnectRooms({ connection: hoveredConnection }));
        }
        editorStore.setSelectedConnection(hoveredConnection);
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
        requestRender();
      }

      // Determine mouse world pos
      state.isMouseOver = true;
      state.mouseWorldX = (event.offsetX - state.cameraOffsetX) / state.cameraZoom;
      state.mouseWorldY = (event.offsetY - state.cameraOffsetY) / state.cameraZoom;

      updateHover();
    };

    // MOUSE LEAVE
    const onMouseLeave = () => {
      stateRef.current.isMouseOver = false;
      updateHover();
    };

    canvas.addEventListener('contextmenu', onContextMenu);
    canvas.addEventListener('click', onMouseClick);
    canvas.addEventListener('mousemove', onMouseMove);
    canvas.addEventListener('mouseleave', onMouseLeave);
    canvas.addEventListener('wheel', onMouseScroll);

    return () => {
      requestRenderRef.current = null;
      cancelAnimationFrame(animationFrameId);
      unsubscribe();
      resizeObserver.disconnect();
      canvas.removeEventListener('contextmenu', onContextMenu);
      canvas.removeEventListener('click', onMouseClick);
      canvas.removeEventListener('mousemove', onMouseMove);
      canvas.removeEventListener('mouseleave', onMouseLeave);
      canvas.removeEventListener('wheel', onMouseScroll);
    };
  }, []);

  return (
    <Box
      ref={canvasBoxRef}
      sx={{
        display: 'block',
        width: '100%',
        height: '100%',
        minWidth: 0,
        overflow: 'hidden',
        backgroundColor: 'background.default',
      }}
    >
      <canvas
        ref={canvasRef}
        style={{
          display: 'block',
          width: '100%',
          height: '100%',
          imageRendering: 'pixelated',
        }}
      />
    </Box>
  )
}

function render(context: CanvasRenderingContext2D, canvasState: CanvasState, colors: CanvasColors) {
  const rooms = editorStore.getRooms();
  const connections = editorStore.getConnections();
  const selectedConnection = editorStore.getSelectedConnection();
  const selectedKey = editorStore.getSelectedCell()?.toString();
  const startKey = editorStore.getStartCell()?.toString();

  context.fillStyle = colors.background;
  context.fillRect(0, 0, context.canvas.width, context.canvas.height);

  context.save();
  context.setTransform(canvasState.cameraZoom, 0, 0, canvasState.cameraZoom, canvasState.cameraOffsetX, canvasState.cameraOffsetY);

  context.strokeStyle = colors.foreground;
  context.fillStyle = colors.foreground;
  context.setLineDash([]);
  context.font = ROOM_FONT;
  context.textAlign = 'center';
  context.textBaseline = 'middle';

  // Only visit the cells that are on screen
  const visibleCells = getVisibleCellRange(canvasState, context.canvas.width, context.canvas.height);

  // Render rooms, collecting their connections so they are drawn after all the rooms
  const visibleConnections: EditorConnection[] = [];
  for (let y = visibleCells.minY; y <= visibleCells.maxY; y++) {
    for (let x = visibleCells.minX; x <= visibleCells.maxX; x++) {
      // Matches EditorCell.toString(), without allocating a cell for every empty spot
      const key = `${x},${y}`;
      const room = rooms.get(key);
      if (!room) {
        continue;
      }

      renderRoom(context, {
        x: x,
        y: y,
        renderStyle: key === selectedKey ? RenderStyle.SELECTED : RenderStyle.SOLID,
        text: room.Name,
        isStartRoom: key === startKey,
      });

      const cellConnections = connections.get(key);
      if (!cellConnections || cellConnections.length === 0) {
        continue;
      }

      const cell = new EditorCell(x, y);
      for (const connKey of cellConnections) {
        const connCell = EditorCell.fromString(connKey);

        // Only render the connection if cell is less than connCell
        // This ensures that connections are only rendered once
        if (!cell.isLessThanOrEqualTo(connCell)) {
          continue;
        }

        visibleConnections.push(new EditorConnection(cell, connCell));
      }
    }
  }

  // Render connections
  for (const connection of visibleConnections) {
    const isSelected = selectedConnection?.isEqualTo(connection);
    renderConnection(context, connection.from, connection.to, isSelected ? RenderStyle.SELECTED : RenderStyle.SOLID);
  }

  // New room hover
  const hoveredCell = getHoveredCell(canvasState);
  if (hoveredCell) {
    const hoveredCellKey = hoveredCell.toString();
    if (!rooms.has(hoveredCellKey)) {
      renderRoom(context, {
        x: hoveredCell.x,
        y: hoveredCell.y,
        renderStyle: RenderStyle.DASHED,
        text: '+',
        isStartRoom: false,
      });
    }
  }

  // Connection hover
  const hoveredConnection = getHoveredConnection(canvasState, rooms);
  if (hoveredConnection && !connectionExists(connections, hoveredConnection)) {
    renderConnection(context, hoveredConnection.from, hoveredConnection.to, RenderStyle.DASHED);
  }

  context.restore();
}

// Expects the line dash, font and text alignment that render() sets up, and leaves them as it found them
function renderRoom(context: CanvasRenderingContext2D, params: RenderRoomParams) {
  const rect = getRoomRect(params.x, params.y);

  if (params.renderStyle === RenderStyle.DASHED) {
    context.setLineDash([7, 2]);
  }
  context.strokeRect(rect.x + 0.5, rect.y + 0.5, rect.width, rect.height);

  if (params.renderStyle === RenderStyle.SELECTED) {
    context.setLineDash([7, 2]);
    context.strokeRect(rect.x - 5.5, rect.y - 5.5, rect.width + 11.5, rect.height + 11.5);
  }
  if (params.renderStyle !== RenderStyle.SOLID) {
    context.setLineDash([]);
  }

  const centerX = rect.x + (rect.width / 2);
  const centerY = rect.y + (rect.height / 2);

  context.fillText(params.text, centerX, centerY);

  if (params.isStartRoom) {
    context.font = START_LABEL_FONT;
    context.textAlign = 'left';
    context.textBaseline = 'top';
    context.fillText('START', rect.x + 6, rect.y + 6);

    context.font = ROOM_FONT;
    context.textAlign = 'center';
    context.textBaseline = 'middle';
  }
}

function renderConnection(context: CanvasRenderingContext2D, fromRoom: EditorCell, toRoom: EditorCell, renderStyle: RenderStyle) {
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

  if (renderStyle === RenderStyle.DASHED) {
    context.setLineDash([7, 2]);
  }
  context.beginPath();
  context.moveTo(startX, startY);
  context.lineTo(endX, endY);
  context.stroke();

  if (renderStyle === RenderStyle.SELECTED) {
    context.setLineDash([7, 2]);

    if (startX === endX) {
      // Left line
      context.beginPath();
      context.moveTo(startX - 5.5, startY);
      context.lineTo(endX - 5.5, endY);
      context.stroke();

      // Right line
      context.beginPath();
      context.moveTo(startX + 5.5, startY);
      context.lineTo(endX + 5.5, endY);
      context.stroke();
    } else {
      // Top line
      context.beginPath();
      context.moveTo(startX, startY - 5.5);
      context.lineTo(endX, endY - 5.5);
      context.stroke();

      // Bottom line
      context.beginPath();
      context.moveTo(startX, startY + 5.5);
      context.lineTo(endX, endY + 5.5);
      context.stroke();
    }
  }

  context.setLineDash([]);
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
  if (!canvasState.isMouseOver) {
    return null;
  }

  const x = Math.floor(canvasState.mouseWorldX / (ROOM_WIDTH + ROOM_X_SPACING));
  const y = Math.floor(canvasState.mouseWorldY / (ROOM_HEIGHT + ROOM_Y_SPACING));
  const rect = getRoomRect(x, y);
  if (rectHasPoint(rect, canvasState.mouseWorldX, canvasState.mouseWorldY)) {
    return new EditorCell(x, y);
  }

  return null;
}

// Returns the range of cells that intersect the viewport, padded by one cell on every side
// so that connections to off-screen rooms and selection outlines at the edges are still drawn
function getVisibleCellRange(canvasState: CanvasState, canvasWidth: number, canvasHeight: number): CellRange {
  const worldLeft = -canvasState.cameraOffsetX / canvasState.cameraZoom;
  const worldTop = -canvasState.cameraOffsetY / canvasState.cameraZoom;
  const worldRight = (canvasWidth - canvasState.cameraOffsetX) / canvasState.cameraZoom;
  const worldBottom = (canvasHeight - canvasState.cameraOffsetY) / canvasState.cameraZoom;

  return {
    minX: Math.floor(worldLeft / (ROOM_WIDTH + ROOM_X_SPACING)) - 1,
    minY: Math.floor(worldTop / (ROOM_HEIGHT + ROOM_Y_SPACING)) - 1,
    maxX: Math.floor(worldRight / (ROOM_WIDTH + ROOM_X_SPACING)) + 1,
    maxY: Math.floor(worldBottom / (ROOM_HEIGHT + ROOM_Y_SPACING)) + 1,
  };
}

function getHoveredConnection(canvasState: CanvasState, rooms: Map<string, EditorRoom>): EditorConnection | null {
  if (!canvasState.isMouseOver) {
    return null;
  }

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

    return new EditorConnection(cell, adjacent);
  }

  // East connection
  if (canvasState.mouseWorldX >= rect.x + rect.width) {
    const adjacent = new EditorCell(x + 1, y);
    if (!rooms.has(adjacent.toString())) {
      return null;
    }

    return new EditorConnection(cell, adjacent);
  }

  // South connection
  if (canvasState.mouseWorldY >= rect.y + rect.height) {
    const adjacent = new EditorCell(x, y + 1);
    if (!rooms.has(adjacent.toString())) {
      return null;
    }

    return new EditorConnection(cell, adjacent);
  }

  // West connection
  if (canvasState.mouseWorldX < rect.x) {
    const adjacent = new EditorCell(x - 1, y);
    if (!rooms.has(adjacent.toString())) {
      return null;
    }

    return new EditorConnection(cell, adjacent);
  }

  return null;
}

// Returns a string that changes whenever the hovered cell or connection changes
function getHoverKey(canvasState: CanvasState, rooms: Map<string, EditorRoom>): string {
  const hoveredCell = getHoveredCell(canvasState);
  if (hoveredCell) {
    return hoveredCell.toString();
  }

  const hoveredConnection = getHoveredConnection(canvasState, rooms);
  if (hoveredConnection) {
    return `${hoveredConnection.from.toString()}|${hoveredConnection.to.toString()}`;
  }

  return '';
}

function connectionExists(connections: Map<string, string[]>, hoveredConnection: EditorConnection) {
  const fromKey = hoveredConnection.from.toString();
  const toKey = hoveredConnection.to.toString();
  return connections.get(fromKey)?.includes(toKey);
}
