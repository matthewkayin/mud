import { useRef, useEffect } from 'react';
import { Box, Paper, Typography } from '@mui/material';
import { EditorState } from './state';

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
    const onMouseClick = (event) => {
    };
    const onMouseMove = (event) => {
      const BUTTON_RIGHT = 2;
      if ((event.buttons & BUTTON_RIGHT) === BUTTON_RIGHT) {
        const state = stateRef.current;
        state.cameraOffset.x -= event.movementX;
        state.cameraOffset.y -= event.movementY;
      }
    };

    canvas.addEventListener('contextmenu', onContextMenu);
    canvas.addEventListener('click', onMouseClick);
    canvas.addEventListener('mousemove', onMouseMove);

    return () => {
      canvas.removeEventListener('contextmenu', onContextMenu);
      canvas.removeEventListener('click', onMouseClick);
      canvas.removeEventListener('mousemove', onMouseMove);
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
        <Typography>Hello</Typography>
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
