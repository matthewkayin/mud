import { useRef, useEffect } from 'react';
import { Box, Paper, Button } from '@mui/material';
import * as mud from '../mud/types';
import { EDITOR_HOVERED_ROOM_GRID_INDEX_NONE, EditorState } from './state';

const CAMERA_ZOOM_MIN = 0.5;
const CAMERA_ZOOM_MAX = 2.0;

export const Editor = () => {
  const canvasRef = useRef(null);
  const canvasBoxRef = useRef(null);
  const stateRef = useRef(new EditorState());

  // RESIZE
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

        stateRef.current.render(context);
      }
    });
    resizeObserver.observe(box);

    return () => resizeObserver.disconnect();
  }, []);

  // CANVAS EVENT LISTENERS
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) {
      return;
    }

    // onContextMenu prevents the default right-click menu from popping up, allowing us
    // to use right-click to drag the camera
    const onContextMenu = (event) => {
      event.preventDefault();
    };

    const onMouseClick = () => {
      const state = stateRef.current;
      const shouldCreateRoom =
        state.hoveredRoomGridIndex !== EDITOR_HOVERED_ROOM_GRID_INDEX_NONE &&
        state.roomGrid[state.hoveredRoomGridIndex] === mud.ROOM_NONE;
      if (shouldCreateRoom) {
        state.createRoom()
      }
    };

    const onMouseMove = (event) => {
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
    };

    const onMouseScroll = (event) => {
      const state = stateRef.current;
      const canvas = canvasRef.current;

      const rect = canvas.getBoundingClientRect();
      const mouseX = event.clientX - rect.left;
      const mouseY = event.clientY - rect.top;

      const worldX = (mouseX - state.cameraOffset.x) / state.cameraZoom;
      const worldY = (mouseY - state.cameraOffset.y) / state.cameraZoom;

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

      state.cameraOffset.x = mouseX - worldX * state.cameraZoom;
      state.cameraOffset.y = mouseY - worldY * state.cameraZoom;
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

  // CANVAS DRAW
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
    const render = () => {
      stateRef.current.render(context);
      animationFrameId = requestAnimationFrame(render);
    };

    // start the render loop
    animationFrameId = requestAnimationFrame(render);

    return () => {
      cancelAnimationFrame(animationFrameId);
    };
  }, []);

  return (
    <Box sx={{
      display: 'flex',
      width: '100vw',
      height: '100vh',
      overflow: 'hidden'
    }}>
      <Paper
        elevation={2}
        sx={{
          width: '25%',
          minWidth: '25%',
          height: '100%',
          borderRadius: 0,
          borderRight: '1px solid',
          borderColor: 'divider',
          p: 2,
          boxSixing: 'border-box',
          overflowY: 'auto',
        }}
      >
        <Button variant="outlined">+ Room</Button>
      </Paper>
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
    </Box>
  );
};
