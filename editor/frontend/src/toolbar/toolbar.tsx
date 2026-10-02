import { useEffect } from 'react';
import { editorStore } from '../store';
import { Box, Stack, IconButton, Divider, useColorScheme } from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';
import UndoIcon from '@mui/icons-material/Undo';
import RedoIcon from '@mui/icons-material/Redo';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import LightModeIcon from '@mui/icons-material/LightMode';

export function Toolbar() {
  const { mode, systemMode, setMode } = useColorScheme();
  const resolvedMode = mode === 'system' ? systemMode : mode;
  const isDarkMode = resolvedMode === 'dark';

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
        backgroundColor: 'background.default',
        borderBottom: '2px solid',
        borderColor: 'divider',
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

      <Box sx={{ flexGrow: 1 }}/>
      <IconButton
        title={isDarkMode ? 'Switch to light theme' : 'Switch to dark theme'}
        aria-label={isDarkMode ? 'Switch to light theme' : 'Switch to dark theme'}
        onClick={() => setMode(isDarkMode ? 'light' : 'dark')}
      >
        {isDarkMode ? <LightModeIcon/> : <DarkModeIcon/>}
      </IconButton>
    </Stack>
  );
}
