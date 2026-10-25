import { useCallback, useEffect, useState } from 'react';
import { editorStore } from '../store';
import {
  Box,
  Stack,
  IconButton,
  Divider,
  useColorScheme,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Button,
} from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';
import UndoIcon from '@mui/icons-material/Undo';
import RedoIcon from '@mui/icons-material/Redo';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import LightModeIcon from '@mui/icons-material/LightMode';
import { LoadWorld, SaveWorld } from '../api/editor_api';

type ErrorDialogState = {
  title: string;
  message: string;
}

export function Toolbar() {
  const { mode, systemMode, setMode } = useColorScheme();
  const resolvedMode = mode === 'system' ? systemMode : mode;
  const isDarkMode = resolvedMode === 'dark';

  const [errorDialog, setErrorDialog] = useState<ErrorDialogState | null>(null);

  // Load the world file on startup
  useEffect(() => {
    LoadWorld()
      .then(editorStore.loadEditorWorld)
      .catch((error) => setErrorDialog({ title: 'Could not open world', message: String(error) }));
  }, []);

  const saveWorld = useCallback(async () => {
    try {
      const savePoint = editorStore.getTopAction();
      await SaveWorld(editorStore.toEditorWorld());
      editorStore.markSaved(savePoint);
    } catch (error) {
      setErrorDialog({ title: 'Could not save world', message: String(error) });
    }
  }, []);

  // Shortcuts
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.ctrlKey && event.code === 'KeyZ') {
        event.preventDefault();
        editorStore.undoAction();
      } else if (event.ctrlKey && event.code === 'KeyR') {
        event.preventDefault();
        editorStore.redoAction();
      } else if (event.ctrlKey && event.code === 'KeyS') {
        event.preventDefault();
        saveWorld();
      }
    };
    window.addEventListener('keydown', onKeyDown);

    return () => {
      window.removeEventListener('keydown', onKeyDown);
    };
  }, [saveWorld]);

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
      <IconButton title="Save (Ctrl+S)" aria-label="Save" onClick={saveWorld}>
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

      <Dialog open={errorDialog !== null} onClose={() => setErrorDialog(null)}>
        <DialogTitle>{errorDialog?.title}</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ whiteSpace: 'pre-line' }}>
            {errorDialog?.message}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setErrorDialog(null)}>OK</Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
