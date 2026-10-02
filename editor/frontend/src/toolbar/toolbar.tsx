import { useEffect } from 'react';
import { editorStore } from '../store';
import { Stack, IconButton, Divider } from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';
import UndoIcon from '@mui/icons-material/Undo';
import RedoIcon from '@mui/icons-material/Redo';

export function Toolbar() {
  // Shortcuts
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.ctrlKey && event.code === 'KeyZ') {
        event.preventDefault();
        editorStore.undoAction();
      } else if (event.ctrlKey && event.code === 'KeyR') {
        event.preventDefault();
        editorStore.redoAction();
      }
    };
    window.addEventListener('keydown', onKeyDown);

    return () => {
      window.removeEventListener('keydown', onKeyDown);
    };
  }, []);

  return (
    <Stack
      direction="row"
      spacing={2}
      sx={{
        width: '100%',
        backgroundColor: '#f0f0f0',
        borderBottom: '2px solid #d6d6d6',
      }}
    >
      <IconButton>
        <SaveIcon/>
      </IconButton>

      <Divider orientation='vertical' flexItem />
      <IconButton onClick={() => editorStore.undoAction()}>
        <UndoIcon/>
      </IconButton>
      <IconButton onClick={() => editorStore.redoAction()}>
        <RedoIcon/>
      </IconButton>
    </Stack>
  );
}
