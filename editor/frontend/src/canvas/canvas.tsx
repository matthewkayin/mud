import { useRef, useEffect, useSyncExternalStore } from 'react';
import { Box } from '@mui/material';
import { editorStore, EditorGrid, ROOM_NONE, type EditorGridCell, editorGridCellToString, editorGridCellFromString } from '../store';
import { world } from '../../wailsjs/go/models';
import { EditorActionAddRoom } from '../store/action';

const CAMERA_ZOOM_MIN = 0.25;
const CAMERA_ZOOM_MAX = 2.0;

const ROOM_X_SPACING = 20;
const ROOM_Y_SPACING = 20;
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

type RenderRoomParams = {
  x: number;
  y: number;
  dashBorder: boolean;
  color: string;
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

  // Resize
  useEffect(() => {
    const canvas = canvasRef.current;
    const box = canvasBoxRef.current;
    if (!canvas || !box) {
      return;
    }

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        canvas.width = width;
        canvas.height = height;

        const context = canvas.getContext('2d');
        if (!context) {
          return;
        }
      }
    });
    resizeObserver.observe(box);

    return () => resizeObserver.disconnect();
  }, []);

  // Event listeners
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) {
      return;
    }

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
      const grid = editorGridRef.current;

      const gridCell = getHoveredGridCell(stateRef.current);
      if (!gridCell) {
        return;
      }

      const cellKey = editorGridCellToString(gridCell);
      const roomIndex = grid.cellToRoomIndex.get(cellKey);

      if (roomIndex === undefined) {
        editorStore.doAction(new EditorActionAddRoom(gridCell));
      } else {
        editorStore.setSelectedGridCell(gridCell);
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

    canvas.addEventListener('contextmenu', onContextMenu);
    canvas.addEventListener('click', onMouseClick);
    canvas.addEventListener('mousemove', onMouseMove);
    canvas.addEventListener('wheel', onMouseScroll);

    return () => {
      canvas.removeEventListener('contextmenu', onContextMenu);
      canvas.removeEventListener('click', onMouseClick);
      canvas.removeEventListener('mousemove', onMouseMove);
      canvas.removeEventListener('wheel', onMouseScroll);
    };
  }, []);

  // Editor grid slice
  const editorGrid = useSyncExternalStore(editorStore.subscribe, () => editorStore.getGrid());
  const editorGridRef = useRef(editorGrid);
  useEffect(() => {
    editorGridRef.current = editorGrid;
  }, [editorGrid]);

  // Editor rooms slice
  const editorRooms = useSyncExternalStore(editorStore.subscribe, () => editorStore.getRooms());
  const editorRoomsRef = useRef(editorRooms);
  useEffect(() => {
    editorRoomsRef.current = editorRooms;
  }, [editorRooms]);

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
      render(context, stateRef.current, editorRoomsRef.current, editorGridRef.current);
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
        width: '100%',
        height: '100%',
        backgroundColor: '#000',
      }}
    >
      <canvas
        ref={canvasRef}
        style={{
          width: '100%',
          height: '100%',
          imageRendering: 'pixelated',
        }}
      />
    </Box>
  )
}

function render(context: CanvasRenderingContext2D, canvasState: CanvasState, rooms: world.Room[], grid: EditorGrid) {
  context.fillStyle = '#000';
  context.fillRect(0, 0, context.canvas.width, context.canvas.height);

  context.save();
  context.setTransform(canvasState.cameraZoom, 0, 0, canvasState.cameraZoom, canvasState.cameraOffsetX, canvasState.cameraOffsetY);

  // Render grid
  for (const [cellKey, roomIndex] of grid.cellToRoomIndex) {
    const cell = editorGridCellFromString(cellKey);
    const isSelected = cellKey === grid.selectedCellKey
    renderRoom(context, {
      x: cell.x,
      y: cell.y,
      dashBorder: false,
      color: isSelected ? '#ffff00' : '#fff',
      text: rooms[roomIndex].Name,
    });
  }

  // New room hover
  const hoveredCell = getHoveredGridCell(canvasState);
  if (hoveredCell) {
    const hoveredCellKey = editorGridCellToString(hoveredCell);
    const hoveredRoomIndex = grid.cellToRoomIndex.get(hoveredCellKey);
    if (hoveredRoomIndex === undefined) {
      renderRoom(context, {
        x: hoveredCell.x,
        y: hoveredCell.y,
        dashBorder: true,
        color: '#fff',
        text: '+',
      });
    }
  }

  context.restore();
}

function renderRoom(context: CanvasRenderingContext2D, params: RenderRoomParams) {
  context.strokeStyle = params.color;
  context.fillStyle = params.color;

  const rect = getRoomRect(params.x, params.y);

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

function getHoveredGridCell(canvasState: CanvasState): EditorGridCell | undefined {
  const cell = {
    x: Math.floor(canvasState.mouseWorldX / (ROOM_WIDTH + ROOM_X_SPACING)),
    y: Math.floor(canvasState.mouseWorldY / (ROOM_HEIGHT + ROOM_Y_SPACING)),
  };
  const rect = getRoomRect(cell.x, cell.y);
  if (!rectHasPoint(rect, canvasState.mouseWorldX, canvasState.mouseWorldY)) {
    return undefined;
  }

  return cell;
}
