import { useRef, useEffect, useState, useCallback } from 'react';
import { Box } from '@mui/material';
import * as mud from '../mud/types';
import { apiWorldGetData } from './api';
import { Sidebar } from './sidebar/sidebar';
import { Canvas } from './canvas/canvas';
import { EDITOR_GRID_INDEX_NONE, EditorState, EditorActionType } from './state';
import type { CanvasMouseWheelEvent as CanvasMouseScrollEvent } from './canvas/canvas';

const CAMERA_ZOOM_MIN = 0.5;
const CAMERA_ZOOM_MAX = 2.0;

export function Editor() {
  const [sidebarGeneration, setSidebarGeneration] = useState(0);
  const [itemData, setItemData] = useState<mud.ItemData[]>([]);
  const stateRef = useRef(new EditorState());

  // Set state setSidebarGeneration
  useEffect(() => {
    stateRef.current.setSidebarGeneration = setSidebarGeneration;
  }, [setSidebarGeneration]);

  // Load item data
  useEffect(() => {
    const getData = async () => {
      const data = await apiWorldGetData('/api/world/items');
      setItemData(data as mud.ItemData[]);
    };
    getData();
  }, []);

  // Canvas callbacks

  const onRender = useCallback((context: CanvasRenderingContext2D) => {
    stateRef.current.render(context);
  }, []);

  const onMouseClick = useCallback(() => {
    const state = stateRef.current;

    // Create room
    const shouldCreateRoom =
      state.hoveredRoomGridIndex !== EDITOR_GRID_INDEX_NONE &&
      state.roomGrid[state.hoveredRoomGridIndex] === mud.ROOM_NONE;
    if (shouldCreateRoom) {
      state.doAction({
        type: EditorActionType.ADD_ROOM,
        data: {
          roomGridIndex: state.hoveredRoomGridIndex,
        },
      });
    }

    // Select the room that was clicked
    const hoveredRoomIndex = state.roomGrid[state.hoveredRoomGridIndex];
    if (hoveredRoomIndex !== mud.ROOM_NONE) {
      state.selectedRoomGridIndex = state.hoveredRoomGridIndex;
      state.refreshSidebar();
    }
  }, []);

  const onMouseMove = useCallback((event: MouseEvent) => {
    const BUTTON_RIGHT = 2;
    const state = stateRef.current;

    // Mouse drag
    if ((event.buttons & BUTTON_RIGHT) === BUTTON_RIGHT) {
      state.cameraOffset.x += event.movementX;
      state.cameraOffset.y += event.movementY;
    }

    // Determine mouse world pos

    const mouseWorldPos = {
      x: (event.offsetX - state.cameraOffset.x) / state.cameraZoom,
      y: (event.offsetY - state.cameraOffset.y) / state.cameraZoom,
    };
    state.onMouseMoved(mouseWorldPos);
  }, []);

  const onMouseScroll = useCallback((event: CanvasMouseScrollEvent) => {
    const state = stateRef.current;

    const worldX = (event.mouseX - state.cameraOffset.x) / state.cameraZoom;
    const worldY = (event.mouseY - state.cameraOffset.y) / state.cameraZoom;

    if (event.scrollY < 0) {
      state.cameraZoom += state.cameraZoom * 0.1;
    } else if (event.scrollY > 0) {
      state.cameraZoom -= state.cameraZoom * 0.1;
    }

    if (state.cameraZoom < CAMERA_ZOOM_MIN) {
      state.cameraZoom = CAMERA_ZOOM_MIN;
    } else if (state.cameraZoom > CAMERA_ZOOM_MAX) {
      state.cameraZoom = CAMERA_ZOOM_MAX;
    }

    state.cameraOffset.x = event.mouseX - worldX * state.cameraZoom;
    state.cameraOffset.y = event.mouseY - worldY * state.cameraZoom;
  }, []);

  return (
    <Box sx={{
      display: 'flex',
      width: '100vw',
      height: '100vh',
      overflow: 'hidden'
    }}>
      <Sidebar
        generation={sidebarGeneration}
        stateRef={stateRef}
        itemData={itemData}
      />
      <Canvas
        onRender={onRender}
        onMouseClick={onMouseClick}
        onMouseMove={onMouseMove}
        onMouseScroll={onMouseScroll}
      />
    </Box>
  );
};
