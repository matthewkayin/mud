import { useEffect, useRef } from 'react';
import { Box } from '@mui/material';

export type CanvasMouseWheelEvent = {
  mouseX: number;
  mouseY: number;
  scrollY: number;
}

export type CanvasProps = {
  onRender: (context: CanvasRenderingContext2D) => void;
  onMouseClick: () => void;
  onMouseMove: (event: MouseEvent) => void;
  onMouseScroll: (event: CanvasMouseWheelEvent) => void;
}

export function Canvas({ onRender, onMouseClick, onMouseMove, onMouseScroll }: CanvasProps) {
  const canvasRef = useRef(null);
  const canvasBoxRef = useRef(null);

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

        onRender(context);
      }
    });
    resizeObserver.observe(box);

    return () => resizeObserver.disconnect();
  }, [onRender]);

  // Event Listeners
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

    const onMouseScrollInternal = (event: WheelEvent) => {
      const canvas = canvasRef.current;
      const canvasRect = canvas.getBoundingClientRect();

      onMouseScroll({
        mouseX: event.clientX - canvasRect.left,
        mouseY: event.clientY - canvasRect.top,
        scrollY: event.deltaY,
      });
    };

    canvas.addEventListener('contextmenu', onContextMenu);
    canvas.addEventListener('click', onMouseClick);
    canvas.addEventListener('mousemove', onMouseMove);
    canvas.addEventListener('wheel', onMouseScrollInternal);

    return () => {
      canvas.removeEventListener('contextmenu', onContextMenu);
      canvas.removeEventListener('click', onMouseClick);
      canvas.removeEventListener('mousemove', onMouseMove);
      canvas.removeEventListener('wheel', onMouseScroll);
    };
  }, [onMouseClick, onMouseMove, onMouseScroll]);

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
    const render = () => {
      onRender(context);
      animationFrameId = requestAnimationFrame(render);
    };

    // start the render loop
    animationFrameId = requestAnimationFrame(render);

    return () => {
      cancelAnimationFrame(animationFrameId);
    };
  }, [onRender]);

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
